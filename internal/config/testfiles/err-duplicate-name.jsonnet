// Invalid because two backups have the same name.
{
  version: 'v1alpha1',
  backups: [
    {
      name: 'backup1',
      interval: '1h',
      commands: [
        {
          cmd: 'echo',
          args: ['foo'],
        },
      ],
    },
    {
      name: 'backup1',
      interval: '2h',
      commands: [
        {
          cmd: 'echo',
          args: ['bar'],
        },
      ],
    },
  ],
}
