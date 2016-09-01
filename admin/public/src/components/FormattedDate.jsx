export function formatDate(date) {
  if (!date)
    return ''

  let parsedDate = new Date(date * 1000)
  return `
    ${parsedDate.getUTCHours()}:${parsedDate.getUTCMinutes()} \
    ${parsedDate.getUTCMonth()}/${parsedDate.getUTCDay()}/${parsedDate.getUTCFullYear()} \
    UTC
  `
}
