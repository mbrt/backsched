// Invalid because the version 'v2' is unknown to backsched.
{
  version: 'v2',
  backups: [
    {
      name: 'rsync-some',
      interval: '1h',
      commands: [
        {
          cmd: 'echo',
          args: ['foo'],
        },
      ],
    },
  ],
}
