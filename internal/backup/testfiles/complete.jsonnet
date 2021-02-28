local days(x) = (24 * x) + 'h';

local backupcmd(name) = [
  {
    cmd: 'echo',
    args: ['start', name],
    env: {
      env1: 'val1',
    },
    workdir: '/home',
  },
  {
    cmd: 'echo',
    args: ['stop', name],
  },
];

{
  version: 'v1alpha1',
  backups: [
    {
      name: 'weekly',
      interval: days(7),
      requires: [
        {
          path: '/mnt/backup/dir1',
        },
        {
          path: '/mnt/backup/dir2',
        },
      ],
      commands: backupcmd('weekly'),
    },
    {
      name: 'hourly',
      interval: '1h',
      requires: [
        {
          path: '/mnt/backup/dir2',
        },
      ],
      commands: backupcmd('hourly'),
    },
  ],
}
