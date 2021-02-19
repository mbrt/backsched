local days(x) = (24 * x) + 'h';

{
  version: 'v1',
  backups: [
    {
      name: 'foo',
      interval: days(7),
      commands: [
        {
          cmd: 'echo',
          args: [
            'start',
          ],
        },
        {
          cmd: 'sleep',
          args: [
            '60',
          ],
        },
        {
          cmd: 'echo',
          args: [
            'stop',
          ],
        },
      ],
    },
  ],
}
