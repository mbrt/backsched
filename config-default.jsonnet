local days(x) = (24 * x) + 'h';

{
  version: 'v1alpha1',
  backups: [
    {
      name: 'weekly',
      interval: days(7),
      commands: [
        {
          cmd: 'echo',
          args: [
            'start',
          ],
        },
      ],
    },
  ],
}
