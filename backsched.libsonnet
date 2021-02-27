// backsched standard library.

local ensureDir(p) = if std.endsWith(p, '/') then p else p + '/';
local joinPath(p, q) = ensureDir(p) + q;


{
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

  restic(src, dest, subdirs, gcloud_cfg)::
    [
      {
        cmd: 'restic',
        args: [
          '-r',
          dest,
          'backup',
        ] + subdirs,
        env: {
          HOME: std.extVar('HOME'),
        },
        workdir: src,
      },
    ],
}
