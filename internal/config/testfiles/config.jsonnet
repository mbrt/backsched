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
        keep_last=10,
        gcloud={
          project_id: 'backup-123456',
          creds_path: lib.env.HOME + '/key.json',
        },
      ),
    },
  ],
}
