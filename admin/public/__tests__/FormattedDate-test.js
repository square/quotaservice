import { formatDate } from '../src/components/FormattedDate';

test('date formatting', () => {
  expect(formatDate(1487202496)).toEqual(`
    23:48
    02/15/2017
    UTC
  `)
})
