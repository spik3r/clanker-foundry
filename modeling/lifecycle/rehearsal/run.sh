#!/usr/bin/env bash
# Fixed-fixture local demonstration; this is not a sandbox or approval service.
set -euo pipefail
if [ "$#" -ne 0 ]; then printf '%s\n' 'rehearsal: no arguments accepted' >&2; exit 2; fi
module="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd -P)"
exec python3 - "$module" <<'PY'
import base64
import hashlib
import io
import json
import os
import pathlib
import shutil
import subprocess
import sys
import tempfile
import zipfile

# The expected fixture was separately authored and frozen before this script.
EXPECTED_SHA256 = '393b2080e0180581fff49857b08a5f9f86314447ee58211e7acf88b67bc0206f'
source = pathlib.Path(sys.argv[1])
# Keep only documented local execution settings; do not inherit arbitrary Go
# flags, tool overrides or proxy configuration from the calling environment.
env = {key: os.environ[key] for key in ('PATH', 'HOME', 'TMPDIR', 'GOCACHE', 'LANG', 'LC_ALL') if key in os.environ}
env.update(GOTOOLCHAIN='local', GOPROXY='off', GOSUMDB='off', GOWORK='off', GOENV='off',
           GOFLAGS='-buildvcs=false', CGO_ENABLED='0')

class RehearsalFailure(Exception):
    pass

def require(condition, stage):
    if not condition:
        raise RehearsalFailure(stage)

def invoke(stage, args, cwd, data=None, code=0):
    try:
        p = subprocess.run(args, input=data, stdout=subprocess.PIPE, stderr=subprocess.PIPE,
                           cwd=cwd, env=env, timeout=180, check=False)
    except (OSError, subprocess.TimeoutExpired):
        raise RehearsalFailure(stage) from None
    require(p.returncode == code, stage)
    return p.stdout

def encoded(value):
    return json.dumps(value, ensure_ascii=False, separators=(',', ':')).encode('utf-8')

def parse(data, stage):
    try:
        return json.loads(data)
    except (ValueError, UnicodeDecodeError):
        raise RehearsalFailure(stage) from None

def sha(data):
    return hashlib.sha256(data).hexdigest()

def b64(data):
    return base64.b64encode(data).decode('ascii')

def check_review(review, expected, stage):
    require(review['mode'] == 'example-only' and review['authority'] == 'none', stage)
    require(review['beforeDigest'] == expected['externalExpected']['BeforeModelDigest'] and
            review['afterDigest'] == expected['externalExpected']['CandidateModelDigest'], stage)
    require(review['changed'] is True and review['reviewRequired'] is True and
            review['changes'] == expected['changes'], stage)
    links = [{'side': x['side'], 'transitionId': x['id'], 'causes': x['causes']}
             for x in review['linkedTransitions']]
    require(links == expected['linkedTransitions'] and review['bindings'] == [], stage)

def check_replay(report, expected, side):
    facts = expected[side]
    expected_model = expected['externalExpected']['BeforeModelDigest' if side == 'baseline' else 'CandidateModelDigest']
    expected_suite = expected['baselineSuiteDigest' if side == 'baseline' else 'candidateSuiteDigest']
    expected_source = expected['baselineSuiteSHA256'] if side == 'baseline' else expected['externalExpected']['SuiteSHA256']
    require(report['mode'] == 'example-only' and report['authority'] == 'none' and
            report['modelDigest'] == expected_model and report['scenarioDigest'] == expected_suite and
            report['sourceDigest'] == expected_source and report['matched'] is True and
            report['casesCompleted'] == facts['cases'] and report['stepsExecuted'] == facts['steps'], side + ' replay')
    entries = report['entries']
    case_order = []
    for entry in entries:
        if not case_order or case_order[-1] != entry['caseId']:
            case_order.append(entry['caseId'])
    require(case_order == facts['caseOrder'], side + ' replay order')
    actuals = []
    for entry in entries:
        actual = entry['actual']
        require(entry['matched'] is True and entry['differences'] == [] and
                actual['mode'] == 'example-only' and actual['authority'] == 'none' and
                actual['modelDigest'] == expected_model, side + ' replay envelope')
        actuals.append({'caseId': entry['caseId'], 'stepId': entry['stepId'], 'ok': actual['ok'],
                        'transitionId': actual['transitionId'], 'task': actual['task'], 'refusals': actual['refusals']})
    require(actuals == expected[side + 'ExpectedReplay'], side + ' replay literals')

try:
    go = shutil.which('go')
    require(go is not None, 'Go toolchain unavailable')
    version = invoke('toolchain', [go, 'env', 'GOVERSION'], source).strip()
    require(version == b'go1.27.3', 'toolchain')
    # A newly allocated private directory is the only location modified. This is
    # not a sandbox against a hostile host or a child-process containment system.
    with tempfile.TemporaryDirectory(prefix='foundry-design-rehearsal-') as tmp:
        work = pathlib.Path(tmp)
        module = work / 'lifecycle'
        shutil.copytree(source, module)
        expected_bytes = (module / 'rehearsal/testdata/expected.json').read_bytes()
        require(sha(expected_bytes) == EXPECTED_SHA256, 'frozen oracle identity')
        expected = parse(expected_bytes, 'frozen oracle')
        baseline_bytes = (module / 'rehearsal/testdata/baseline-suite.json').read_bytes()
        candidate_suite_bytes = (module / 'designchange/testdata/suite.json').read_bytes()
        request_bytes = (module / 'designchange/testdata/request.json').read_bytes()
        require(sha(baseline_bytes) == expected['baselineSuiteSHA256'] and
                sha(candidate_suite_bytes) == expected['externalExpected']['SuiteSHA256'] and
                sha(request_bytes) == expected['externalExpected']['RequestSHA256'], 'fixture identities')
        patch_request = parse(request_bytes, 'patch fixture')
        baseline_suite = parse(baseline_bytes, 'baseline suite')
        candidate_suite = parse(candidate_suite_bytes, 'candidate suite')
        before = patch_request['before']
        bins = work / 'bin'
        bins.mkdir()
        targets = ['modelpatch', 'modeldiff', 'modeltopology', 'modelscenariodiff', 'modeldesignchange', 'modelplay', 'modelc']
        invoke('build baseline tools', [go, 'build', '-o', str(bins) + os.sep] + ['./cmd/' + x for x in targets], module)
        tool = lambda name: str(bins / name)

        baseline_report = parse(invoke('baseline replay', [tool('modelplay')], module, baseline_bytes), 'baseline replay')
        check_replay(baseline_report, expected, 'baseline')
        require(invoke('candidate binding refusal', [tool('modelplay')], module, candidate_suite_bytes, 2) == b'', 'candidate binding refusal')

        plan = parse(invoke('patch planning', [tool('modelpatch')], module, request_bytes), 'patch planning')
        require(plan['reportVersion'] == 'lifecycle-model-patch-report/v1' and plan['profileVersion'] == 'lifecycle/v1' and
                plan['mode'] == 'example-only' and plan['authority'] == 'none' and
                plan['beforeDigest'] == expected['externalExpected']['BeforeModelDigest'] and
                plan['afterDigest'] == expected['externalExpected']['CandidateModelDigest'], 'patch identity')
        check_review(plan['review'], expected, 'patch review')

        # Package the candidate and original suite using the existing exact source
        # policy, then verify against separately frozen external identities.
        names = ['go.mod', 'model/ir.go', 'model/capabilities.go', 'canonical/canonical.go',
                 'strictjson/decode.go', 'compiler/generate.go', 'scenarios/types.go',
                 'scenarios/decode.go', 'scenarios/run.go', 'scenarios/runner.go',
                 'scenarios/report_decode.go', 'cmd/modelplay/main.go']
        carrier = {'formatVersion': 'lifecycle-designchange-json/v1', 'operation': 'build',
                   'requestBase64': b64(request_bytes), 'suiteBase64': b64(candidate_suite_bytes),
                   'sources': [{'name': name, 'dataBase64': b64((module / name).read_bytes())} for name in names],
                   'archiveBase64': None, 'expected': None}
        archive = invoke('design-change export', [tool('modeldesignchange')], module, encoded(carrier))
        external = {key[0].lower() + key[1:]: value for key, value in expected['externalExpected'].items()}
        carrier = {'formatVersion': 'lifecycle-designchange-json/v1', 'operation': 'verify',
                   'requestBase64': None, 'suiteBase64': None, 'sources': None,
                   'archiveBase64': b64(archive), 'expected': external}
        verification = parse(invoke('design-change verification', [tool('modeldesignchange')], module, encoded(carrier)), 'design-change verification')
        require(verification == {'formatVersion': 'lifecycle-designchange-verification/v1', 'consistent': True,
                'mode': 'example-only', 'authority': 'none', 'suiteStatus': expected['suiteStatus'],
                'suiteBindingStatus': expected['suiteBindingStatus']}, 'design-change verification')
        # Verification precedes reading this exact member. No filesystem archive
        # extraction occurs; the independently frozen hash binds the selected bytes.
        with zipfile.ZipFile(io.BytesIO(archive)) as z:
            members = [x for x in z.infolist() if x.filename == 'candidate.json']
            require(len(members) == 1 and members[0].file_size <= 1 << 20, 'candidate archive member')
            candidate_bytes = z.read(members[0])
        require(sha(candidate_bytes) == expected['externalExpected']['CandidateSHA256'], 'candidate byte identity')
        candidate = parse(candidate_bytes, 'candidate model')
        require(plan['after'] == candidate, 'planner and archive candidate')

        review_request = {'formatVersion': 'lifecycle-model-diff-json/v1', 'before': before, 'after': candidate, 'bindings': []}
        review = parse(invoke('model review', [tool('modeldiff')], module, encoded(review_request)), 'model review')
        check_review(review, expected, 'model review')
        suite_request = {'formatVersion': 'lifecycle-scenario-diff-json/v1', 'beforeModel': before,
                         'beforeSuite': baseline_suite, 'afterModel': candidate, 'afterSuite': candidate_suite}
        suite_request_bytes = encoded(suite_request)
        suite_review = parse(invoke('suite review', [tool('modelscenariodiff')], module, suite_request_bytes), 'suite review')
        # These wire hashes identify the newly assembled JSON carrier and its
        # embedded tokens, not the original files or authenticated provenance.
        require(suite_review['sources'] == {'request': sha(suite_request_bytes),
                'beforeModel': sha(encoded(before)), 'beforeSuite': sha(encoded(baseline_suite)),
                'afterModel': sha(encoded(candidate)), 'afterSuite': sha(encoded(candidate_suite))}, 'suite review sources')
        suite_review['sources'] = None
        require(suite_review == expected['suiteDiff'], 'suite review literals')
        for model_value, digest in [(before, expected['externalExpected']['BeforeModelDigest']),
                                    (candidate, expected['externalExpected']['CandidateModelDigest'])]:
            topology_request = {'formatVersion': 'lifecycle-structural-topology-request/v1', 'model': model_value}
            topology = parse(invoke('topology', [tool('modeltopology')], module, encoded(topology_request)), 'topology')
            expected_topology = dict(expected['sharedStructuralTopology'], modelDigest=digest)
            require(topology == expected_topology, 'structural topology literals')

        # Only this private copy is changed. Rebuild a runner from the candidate;
        # neither old nor new suite is ever edited or rebound by the workflow.
        (module / 'examples/lifecycle.json').write_bytes(candidate_bytes)
        invoke('candidate generation', [tool('modelc'), 'generate'], module)
        invoke('candidate generation check', [tool('modelc'), 'check'], module)
        generated_names = ['generated/lifecycle.go', 'generated/model_paths_test.go', 'generated/decision.svg',
                           'generated/decision-table.json', 'generated/trace.json']
        artifact_hashes = {name: sha((module / name).read_bytes()) for name in generated_names}
        invoke('generated structural tests', [go, 'test', '-count=1', './generated'], module)
        candidate_runner = str(bins / 'modelplay-candidate')
        invoke('build candidate runner', [go, 'build', '-o', candidate_runner, './cmd/modelplay'], module)
        require(invoke('baseline binding refusal', [candidate_runner], module, baseline_bytes, 2) == b'', 'baseline binding refusal')
        candidate_report = parse(invoke('candidate replay', [candidate_runner], module, candidate_suite_bytes), 'candidate replay')
        check_replay(candidate_report, expected, 'candidate')
        result = {'formatVersion': 'lifecycle-design-rehearsal-result/v1', 'mode': 'example-only', 'authority': 'none',
                  'beforeModelDigest': expected['externalExpected']['BeforeModelDigest'],
                  'candidateModelDigest': expected['externalExpected']['CandidateModelDigest'],
                  'baselineMatched': True, 'baselineSteps': expected['baseline']['steps'],
                  'candidateMatched': True, 'candidateSteps': expected['candidate']['steps'],
                  'candidateBindingRefusedByBaseline': True, 'baselineBindingRefusedByCandidate': True,
                  'modelReviewRequired': True, 'suiteReviewRequired': True, 'topologyMode': 'structural-only',
                  'archiveConsistent': True, 'archiveSuiteStatus': expected['suiteStatus'],
                  'archiveBindingStatus': expected['suiteBindingStatus'], 'archiveSHA256': sha(archive),
                  'generatedArtifactSHA256': artifact_hashes}
    # Publish success only after the private directory has been cleaned up.
    sys.stdout.buffer.write(encoded(result) + b'\n')
except RehearsalFailure as exc:
    # Only fixed stage names are raised by this script; no caller content is echoed.
    print('rehearsal: failed at ' + str(exc), file=sys.stderr)
    sys.exit(1)
except (OSError, KeyError, TypeError, ValueError, RuntimeError, zipfile.BadZipFile):
    print('rehearsal: local operation failed', file=sys.stderr)
    sys.exit(1)
PY
