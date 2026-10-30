import { createHash } from 'node:crypto';
import Ajv2020 from 'ajv/dist/2020.js';
import schema from '../contracts/v1.schema.json' with { type: 'json' };
import authoritySchema from '../contracts/authority.schema.json' with { type: 'json' };
import evidenceSchema from '../contracts/evidence.schema.json' with { type: 'json' };
import workspaceSchema from '../contracts/workspace.schema.json' with { type: 'json' };

const ajv = new Ajv2020({
  allErrors: true,
  strict: true,
  strictTypes: true,
  coerceTypes: false,
  useDefaults: false,
  removeAdditional: false,
});

const schemas = [schema, authoritySchema, evidenceSchema];
for (const contract of schemas) ajv.addSchema(contract);
ajv.addSchema(workspaceSchema);

const validators = new Map(
  schemas.flatMap((contract) => Object.keys(contract.$defs).map((type) => [
    type,
    ajv.compile({ $ref: `${contract.$id}#/$defs/${type}` }),
  ])),
);
validators.set('WorkspacePlan', ajv.compile({ $ref: workspaceSchema.$id }));

const forbiddenKeys = new Set(['__proto__', 'constructor', 'prototype']);

function contractError(code, message, details = []) {
  const error = new Error(message);
  error.code = code;
  error.details = details;
  return error;
}

/** Validate a named v1 wire record and return the original value unchanged. */
export function validate(type, value) {
  const validator = validators.get(type);
  if (!validator) {
    throw contractError('UNKNOWN_CONTRACT', `Unknown contract type: ${type}`);
  }
  // Ajv's object-array equality assumes Object.prototype methods. Keep the public
  // parser's null-prototype objects, but validate a safe JSON-normalized copy.
  const input = JSON.parse(canonical(value));
  let valid;
  try {
    valid = validator(input);
  } catch {
    throw contractError('INVALID_CONTRACT', `${type} contains unsupported comparison data`);
  }
  if (!valid) {
    const details = (validator.errors ?? []).map((item) => ({
      path: item.instancePath,
      keyword: item.keyword,
      message: item.message,
      params: item.params,
    }));
    throw contractError('INVALID_CONTRACT', `${type} does not match its contract`, details);
  }
  return value;
}

function assertUnicode(value) {
  for (let i = 0; i < value.length; i += 1) {
    const code = value.charCodeAt(i);
    if (code >= 0xd800 && code <= 0xdbff) {
      const next = value.charCodeAt(i + 1);
      if (!(next >= 0xdc00 && next <= 0xdfff)) {
        throw contractError('INVALID_UNICODE', 'Unpaired high surrogate in JSON string');
      }
      i += 1;
    } else if (code >= 0xdc00 && code <= 0xdfff) {
      throw contractError('INVALID_UNICODE', 'Unpaired low surrogate in JSON string');
    }
  }
}

/** Parse JSON while detecting duplicate keys and enforcing byte and depth limits. */
export function parseJSON(text, { maxBytes = 1_048_576, maxDepth = 64 } = {}) {
  if (typeof text !== 'string') {
    throw contractError('INVALID_JSON_INPUT', 'JSON input must be a string');
  }
  if (!Number.isSafeInteger(maxBytes) || maxBytes < 1 || !Number.isSafeInteger(maxDepth) || maxDepth < 1) {
    throw contractError('INVALID_PARSE_LIMIT', 'maxBytes and maxDepth must be positive safe integers');
  }
  if (Buffer.byteLength(text, 'utf8') > maxBytes) {
    throw contractError('JSON_TOO_LARGE', `JSON input exceeds ${maxBytes} bytes`);
  }

  let offset = 0;
  const fail = (message) => {
    throw contractError('MALFORMED_JSON', `${message} at byte offset ${Buffer.byteLength(text.slice(0, offset), 'utf8')}`);
  };
  const whitespace = () => {
    while (text[offset] === ' ' || text[offset] === '\n' || text[offset] === '\r' || text[offset] === '\t') offset += 1;
  };
  const string = () => {
    const start = offset;
    if (text[offset] !== '"') fail('Expected string');
    offset += 1;
    while (offset < text.length) {
      const char = text[offset];
      if (char === '"') {
        offset += 1;
        let parsed;
        try {
          parsed = JSON.parse(text.slice(start, offset));
        } catch {
          fail('Invalid string escape');
        }
        assertUnicode(parsed);
        return parsed;
      }
      if (char === '\\') {
        offset += 1;
        if (offset >= text.length) fail('Unterminated escape');
      } else if (char.charCodeAt(0) < 0x20) {
        fail('Unescaped control character');
      }
      offset += 1;
    }
    fail('Unterminated string');
  };
  const value = (depth) => {
    whitespace();
    if (depth > maxDepth) throw contractError('JSON_TOO_DEEP', `JSON nesting exceeds ${maxDepth}`);
    const char = text[offset];
    if (char === '"') return string();
    if (char === '{') {
      offset += 1;
      whitespace();
      const result = Object.create(null);
      const keys = new Set();
      if (text[offset] === '}') {
        offset += 1;
        return result;
      }
      while (offset < text.length) {
        whitespace();
        if (text[offset] !== '"') fail('Expected object key');
        const key = string();
        if (forbiddenKeys.has(key)) throw contractError('UNSAFE_JSON_KEY', `Forbidden object key: ${key}`);
        if (keys.has(key)) throw contractError('DUPLICATE_JSON_KEY', `Duplicate object key: ${key}`);
        keys.add(key);
        whitespace();
        if (text[offset] !== ':') fail('Expected colon');
        offset += 1;
        result[key] = value(depth + 1);
        whitespace();
        if (text[offset] === '}') {
          offset += 1;
          return result;
        }
        if (text[offset] !== ',') fail('Expected comma or closing brace');
        offset += 1;
      }
      fail('Unterminated object');
    }
    if (char === '[') {
      offset += 1;
      whitespace();
      const result = [];
      if (text[offset] === ']') {
        offset += 1;
        return result;
      }
      while (offset < text.length) {
        result.push(value(depth + 1));
        whitespace();
        if (text[offset] === ']') {
          offset += 1;
          return result;
        }
        if (text[offset] !== ',') fail('Expected comma or closing bracket');
        offset += 1;
      }
      fail('Unterminated array');
    }
    for (const [literal, parsed] of [['true', true], ['false', false], ['null', null]]) {
      if (text.startsWith(literal, offset)) {
        offset += literal.length;
        return parsed;
      }
    }
    const match = text.slice(offset).match(/^-?(?:0|[1-9]\d*)(?:\.\d+)?(?:[eE][+-]?\d+)?/);
    if (match) {
      offset += match[0].length;
      const parsed = Number(match[0]);
      if (!Number.isFinite(parsed)) throw contractError('NONFINITE_NUMBER', 'JSON number is outside the finite range');
      return parsed;
    }
    fail('Expected JSON value');
  };

  const parsed = value(0);
  whitespace();
  if (offset !== text.length) fail('Unexpected trailing content');
  return parsed;
}

/** Serialize supported JSON values with lexicographically sorted object keys. */
export function canonical(value) {
  const seen = new Set();
  const encode = (item) => {
    if (item === null) return 'null';
    if (typeof item === 'string') {
      assertUnicode(item);
      return JSON.stringify(item);
    }
    if (typeof item === 'boolean') return item ? 'true' : 'false';
    if (typeof item === 'number') {
      if (!Number.isFinite(item)) throw contractError('UNSUPPORTED_VALUE', 'Canonical JSON does not support non-finite numbers');
      return JSON.stringify(item);
    }
    if (typeof item !== 'object') {
      throw contractError('UNSUPPORTED_VALUE', `Canonical JSON does not support ${typeof item}`);
    }
    if (seen.has(item)) throw contractError('UNSUPPORTED_VALUE', 'Canonical JSON does not support cyclic values');
    seen.add(item);
    let result;
    if (Array.isArray(item)) {
      const ownKeys = Reflect.ownKeys(item);
      if (ownKeys.some((key) => typeof key === 'symbol') || ownKeys.length !== item.length + 1 || !ownKeys.includes('length')) {
        throw contractError('UNSUPPORTED_VALUE', 'Canonical arrays must be dense and have no extra properties');
      }
      const entries = [];
      for (let index = 0; index < item.length; index += 1) {
        const key = String(index);
        const descriptor = Object.getOwnPropertyDescriptor(item, key);
        if (!descriptor || !descriptor.enumerable || !Object.hasOwn(descriptor, 'value')) {
          throw contractError('UNSUPPORTED_VALUE', 'Canonical arrays must contain dense enumerable data elements');
        }
        entries.push(encode(descriptor.value));
      }
      result = `[${entries.join(',')}]`;
    } else {
      const prototype = Object.getPrototypeOf(item);
      if (prototype !== Object.prototype && prototype !== null) {
        throw contractError('UNSUPPORTED_VALUE', 'Canonical objects must be plain objects');
      }
      const keys = Reflect.ownKeys(item);
      if (keys.some((key) => typeof key !== 'string')) {
        throw contractError('UNSUPPORTED_VALUE', 'Canonical objects cannot have symbol keys');
      }
      keys.sort();
      const entries = [];
      for (const key of keys) {
        if (forbiddenKeys.has(key)) throw contractError('UNSAFE_JSON_KEY', `Forbidden object key: ${key}`);
        const descriptor = Object.getOwnPropertyDescriptor(item, key);
        if (!descriptor.enumerable || !Object.hasOwn(descriptor, 'value')) {
          throw contractError('UNSUPPORTED_VALUE', 'Canonical objects must contain enumerable data properties');
        }
        assertUnicode(key);
        entries.push(`${JSON.stringify(key)}:${encode(descriptor.value)}`);
      }
      result = `{${entries.join(',')}}`;
    }
    seen.delete(item);
    return result;
  };
  return encode(value);
}

/** Return the SHA-256 digest of canonical JSON encoded as UTF-8. */
export function digest(value) {
  return createHash('sha256').update(canonical(value), 'utf8').digest('hex');
}
