// backsched standard library.

local ensureDir(p) = if std.endsWith(p, '/') then p else p + '/';

local joinPath(p, q) = ensureDir(p) + q;

local hasFields(o, fs) =
  std.foldl(function(prev, y) prev && std.objectHas(o, y), fs, true);


{
  // env contains some environment variables prefilled by backsched.
  env: {
    HOME: std.extVar('HOME'),
    USER: std.extVar('USER'),
  },

  // rsync.
  //
  // Uses rsync to backup a source to a destination directory.
  // When subdirs is non-empty, rsync is executed multiple times for the
  // given source and destination subdirs: each subdir is appended to both
  // the source and destination root.
  rsync(src, dest, subdirs=[])::
    local run(src, dest) = [
      {
        cmd: 'mkdir',
        args: [
          '--mode=700',
          '-p',
          dest,
        ],
      },
      {
        cmd: 'rsync',
        args: [
          '-av',
          '--delete',
          ensureDir(src),
          ensureDir(dest),
        ],
      },
    ];

    if std.length(subdirs) == 0 then run(src, dest)
    else std.flattenArrays([
      run(joinPath(src, p), joinPath(dest, p))
      for p in subdirs
    ]),

  // restic.
  //
  // Use restic to backup a source with its subdirs into a local destination or
  // to a GCS bucket. If keepLast is given, it specifies the maximum number of
  // snapshots to keep.
  restic(src, dest, subdirs, keepLast=null, gcloud=null)::
    // Check that gcloud has the required args.
    assert gcloud == null || hasFields(gcloud, ['projectId', 'credsPath']) :
           'parameters `projectId` and `credsPath` are required if `gcloud` is not null';
    local env = {
      HOME: $.env.HOME,
      [if gcloud != null then 'GOOGLE_PROJECT_ID']: gcloud.projectId,
      [if gcloud != null then 'GOOGLE_APPLICATION_CREDENTIALS']: gcloud.credsPath,
    };
    local run(args) = {
      cmd: 'restic',
      args: ['-r', dest] + args,
      env: env,
      secretEnv: {
        RESTIC_PASSWORD: {
          id: 'password',
        },
      },
      workdir: src,
    };

    [
      run(['backup', '--one-file-system'] + subdirs),
      run(['check']),
    ] + if keepLast == null then []
    else [
      run([
        'forget',
        '--keep-last',
        std.toString(keepLast),
        '--max-unused=1%',
        '--prune',
      ]),
    ],
}
