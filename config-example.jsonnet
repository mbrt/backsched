local lib = import 'backsched.libsonnet';
local days(x) = (24 * x) + 'h';

{
  version: 'v1alpha1',
  backups: [
    {
      name: 'rsync-some',
      interval: '1h',
      commands: lib.rsync(src='/home/me', dest='/mnt/backup/imp', subdirs=[
        'docs',
        'pics',
        'vids',
      ]),
      requires: [{
        path: '/mnt/backup/imp',
      }],
    },

    {
      name: 'rsync-all',
      interval: days(7),
      commands: lib.rsync(src='/home/me', dest='/mnt/backup/full'),
      requires: [{
        path: '/mnt/backup/full',
      }],
    },

    {
      name: 'restic-some',
      interval: days(7),
      commands: lib.restic(
        src='/home/me',
        dest='gs:personal:/',
        subdirs=['docs', 'pics'],
        keepLast=10,
        gcloud={
          projectId: 'backup-123456',
          credsPath: lib.env.HOME + '/key.json',
        },
      ),
    },
    {
      name: 'restic-local',
      interval: days(3),
      commands: lib.restic(
        src='/home/me',
        dest='/mnt/backup/restic',
        subdirs=['.'],
        keepLast=20,
      ),
      requires: [{
        path: '/mnt/backup/full',
      }],
    },
  ],
}
