"""Fixed local consumer demonstration. No arbitrary request or command interface."""
import hashlib
import json
import os
import pathlib
import shutil
import stat
import subprocess
import sys
import tempfile

ORACLE_SHA256 = '8ae4be53b0ee172073e6e36d9c21f3dbb0e6a39b50e8729d23a85e2669acac4e'
FIXTURES = {'requirements.md', 'model-base.json', 'model-left.json', 'model-right.json',
            'model-merged.json', 'model-conflict.json', 'model-invalid-left.json',
            'model-invalid-right.json', 'suite-base.json', 'suite-left.json',
            'suite-right.json', 'suite-merged.json'}


class Failure(Exception):
    pass


def require(ok, stage):
    if not ok:
        raise Failure(stage)


def sha(raw):
    return hashlib.sha256(raw).hexdigest()


def encoded(value):
    return json.dumps(value, ensure_ascii=False, separators=(',', ':')).encode('utf-8')


def snapshot(root):
    result = {}
    for path in sorted(root.rglob('*')):
        info = path.lstat()
        require(not stat.S_ISLNK(info.st_mode), 'source symlink')
        if stat.S_ISDIR(info.st_mode):
            continue
        require(stat.S_ISREG(info.st_mode), 'source regular files')
        result[str(path.relative_to(root))] = (stat.S_IMODE(info.st_mode), sha(path.read_bytes()))
    return result


class Workflow:
    def __init__(self, source, work, go, env, expected):
        self.source, self.work, self.go, self.env, self.expected = source, work, go, env, expected
        self.module = work / 'lifecycle'
        self.bins = work / 'bin'
        self.generated = 0
        self.replayed = 0

    def invoke(self, stage, args, data=None, code=0):
        try:
            p = subprocess.run(args, input=data, stdout=subprocess.PIPE, stderr=subprocess.PIPE,
                               cwd=self.module, env=self.env, timeout=180, check=False)
        except (OSError, subprocess.TimeoutExpired):
            raise Failure(stage) from None
        require(p.returncode == code, stage)
        return p.stdout

    def build(self, target, package):
        self.invoke('compile ' + package, [self.go, 'build', '-p', '1', '-o', str(target), package])

    def replay_model(self, model_bytes, version):
        # Only the private copy is changed. Existing generation policy remains
        # authoritative, including detection of modified/partial prior outputs.
        (self.module / 'examples/lifecycle.json').write_bytes(model_bytes)
        modelc = str(self.bins / 'modelc')
        self.invoke('generate ' + version, [modelc, 'generate'])
        self.generated += 1
        self.invoke('check generation ' + version, [modelc, 'check'])
        self.invoke('structural tests ' + version,
                    [self.go, 'test', '-p', '1', '-count=1', './generated'])
        runner = self.bins / 'modelplay'
        self.build(runner, './cmd/modelplay')
        suite = (self.source / ('mergerehearsal/testdata/suite-' + version + '.json')).read_bytes()
        raw = self.invoke('replay ' + version, [str(runner)], suite)
        require(json.loads(raw) == self.expected['replays'][version], 'full replay ' + version)
        self.replayed += 1
        stale_version = 'base' if version == 'merged' else 'merged'
        stale = (self.source / ('mergerehearsal/testdata/suite-' + stale_version + '.json')).read_bytes()
        require(self.invoke('stale binding ' + version, [str(runner)], stale, 2) == b'',
                'stale binding output ' + version)
        return {'reportSHA256': sha(raw), 'modelDigest': self.expected['modelDigests'][version],
                'cases': self.expected['replays'][version]['casesCompleted'],
                'steps': self.expected['replays'][version]['stepsExecuted'],
                'generatedGoSHA256': sha((self.module / 'generated/lifecycle.go').read_bytes()),
                'staleSuiteRefused': True}

    def merge_case(self, case):
        raw = self.invoke('merge ' + case['id'], [str(self.bins / 'modelmerge')],
                          case['request'].encode('utf-8'))
        report = json.loads(raw)
        require(report == case['expected'], 'full merge ' + case['id'])
        before = (self.generated, self.replayed)
        result = {'id': case['id'], 'status': report['status'], 'reportSHA256': sha(raw),
                  'generation': 'not-run', 'replay': 'not-run'}
        # Valid reports include conflicts and invalid combinations. Exit 0 alone
        # must never launch generation or replay of a partial candidate.
        if report['status'] != 'merged':
            require(report['candidate'] is None and report['candidateDigest'] is None,
                    'no partial candidate')
            require(before == (self.generated, self.replayed), 'short circuit')
            return result
        candidate = report['candidate']
        require(candidate == json.loads((self.source / 'mergerehearsal/testdata/model-merged.json').read_bytes()),
                'merged candidate literals')
        result.update(generation='observed', replay='observed',
                      candidate=self.replay_model(encoded(candidate), 'merged'))
        return result

    def mutation_controls(self):
        runner = self.bins / 'modelplay'
        suite = (self.source / 'mergerehearsal/testdata/suite-merged.json').read_bytes()
        target = self.module / 'generated/lifecycle.go'
        original = target.read_bytes()
        controls = {}
        for name, guard in [('removed-pack-guard', 'G_packActor'),
                            ('removed-dispatch-guard', 'G_dispatchActor')]:
            needle = ('failed = append(failed, "' + guard + '")').encode('ascii')
            require(original.count(needle) == 1, 'mutation location ' + name)
            try:
                target.write_bytes(original.replace(needle, b'_ = "disposable guard mutation"', 1))
                self.build(runner, './cmd/modelplay')
                raw = self.invoke('mutant mismatch ' + name, [str(runner)], suite, 1)
                require(json.loads(raw) == self.expected['behavioralMutants'][name],
                        'full mutant report ' + name)
                controls[name] = sha(raw)
            finally:
                target.write_bytes(original)
        self.build(runner, './cmd/modelplay')
        restored = self.invoke('restored candidate replay', [str(runner)], suite)
        require(json.loads(restored) == self.expected['replays']['merged'], 'restored full report')
        wrong = self.expected['wrongExpectation']
        wrong_bytes = wrong['suite'].encode('utf-8') if isinstance(wrong['suite'], str) else encoded(wrong['suite'])
        raw = self.invoke('wrong expectation mismatch', [str(runner)], wrong_bytes, 1)
        require(json.loads(raw) == wrong['expected'], 'full wrong-expectation report')
        require(sha(suite) == self.expected['fileSHA256']['suite-merged.json'], 'unchanged expectations')
        controls['wrong-expectation'] = sha(raw)
        controls['restoredReplaySHA256'] = sha(restored)
        return controls

    def run(self):
        shutil.copytree(self.source, self.module)
        self.bins.mkdir(mode=0o700)
        self.build(self.bins / 'modelmerge', './cmd/modelmerge')
        self.build(self.bins / 'modelc', './cmd/modelc')
        reports = [self.merge_case(case) for case in self.expected['merges']]
        # The two successful orientations both compiled/replayed the actual
        # candidate. Restore it after independent base/branch observations.
        revisions = {}
        for version in ('base', 'left', 'right', 'merged'):
            revisions[version] = self.replay_model(
                (self.source / ('mergerehearsal/testdata/model-' + version + '.json')).read_bytes(), version)
        controls = self.mutation_controls()
        return {'formatVersion': 'lifecycle-merge-rehearsal-result/v1', 'mode': 'example-only',
                'authority': 'none', 'oracleSHA256': ORACLE_SHA256, 'merges': reports,
                'revisions': revisions, 'controls': controls,
                'generationRuns': self.generated, 'replayRuns': self.replayed,
                'cleanup': 'completed', 'inputs': 'unchanged'}


def perform(source, env, temporary_directory=tempfile.TemporaryDirectory, workflow=Workflow):
    before = snapshot(source)
    raw = (source / 'mergerehearsal/testdata/expected.json').read_bytes()
    require(sha(raw) == ORACLE_SHA256, 'frozen oracle identity')
    expected = json.loads(raw)
    require(set(expected['fileSHA256']) == FIXTURES, 'fixture inventory')
    for name, digest in expected['fileSHA256'].items():
        require(sha((source / 'mergerehearsal/testdata' / name).read_bytes()) == digest,
                'frozen fixture identity')
    go = shutil.which('go', path=env.get('PATH'))
    require(go is not None, 'Go toolchain unavailable')
    try:
        p = subprocess.run([go, 'env', 'GOVERSION'], stdout=subprocess.PIPE, stderr=subprocess.PIPE,
                           env=env, timeout=30, check=False)
    except (OSError, subprocess.TimeoutExpired):
        raise Failure('Go toolchain unavailable') from None
    require(p.returncode == 0 and p.stdout.strip() == b'go1.27.3', 'exact Go toolchain')
    with temporary_directory(prefix='foundry-merge-rehearsal-') as tmp:
        result = workflow(source, pathlib.Path(tmp), go, env, expected).run()
        require(snapshot(source) == before, 'unchanged checkout bytes and modes')
    # Never return success before context exit confirms temporary cleanup.
    return result


def main():
    if len(sys.argv) != 1:
        print('merge rehearsal: no arguments accepted', file=sys.stderr)
        return 2
    source = pathlib.Path(__file__).resolve().parent.parent
    env = {k: os.environ[k] for k in ('PATH', 'HOME', 'TMPDIR', 'GOTMPDIR', 'GOCACHE',
                                     'LANG', 'LC_ALL', 'GOMAXPROCS') if k in os.environ}
    env.update(GOTOOLCHAIN='local', GOPROXY='off', GOSUMDB='off', GOWORK='off',
               GOENV='off', GOFLAGS='-buildvcs=false', CGO_ENABLED='0')
    try:
        result = perform(source, env)
        sys.stdout.buffer.write(encoded(result) + b'\n')
        sys.stdout.buffer.flush()
        return 0
    except Failure as exc:
        print('merge rehearsal: failed at ' + str(exc), file=sys.stderr)
    except (OSError, ValueError, KeyError, TypeError, RuntimeError):
        print('merge rehearsal: local operation failed', file=sys.stderr)
    return 1


if __name__ == '__main__':
    sys.exit(main())
