// backsched standard library.

local ensureDir(p) = if std.endsWith(p, '/') then p else p + '/';
local joinPath(p, q) = ensureDir(p) + q;


{
  env: {
      HOME: std.extVar('HOME'),
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

  restic(src, dest, subdirs, keep_last=null, gcloud=null)::
    local env = {
      HOME: $.env.HOME,
      GOOGLE_PROJECT_ID: if gcloud != null then gcloud.project_id else null,
      GOOGLE_APPLICATION_CREDENTIALS: if gcloud != null then gcloud.creds_path else null,
    };
    local run(args) = {
      cmd: 'restic',
      args: ['-r', dest] + args,
      env: env,
      workdir: src,
    };

    [
      run(['backup'] + subdirs),
      run(['check']),
    ] + if keep_last == null then []
    else [
      run([
        'forget',
        '--keep-last',
        std.toString(keep_last),
        '--prune',
      ]),
    ],
}
