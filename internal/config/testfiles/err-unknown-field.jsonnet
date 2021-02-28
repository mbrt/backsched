// Invalid because there's an unknown field.
{
  version: 'v1alpha1',
  backups: [
    {
      name: 'rsync-some',
      interval: '1h',
      this_field_should_not_be_here: true,
      commands: [
        {
          cmd: 'echo',
          args: ['foo'],
        },
      ],
    },
  ],
}
