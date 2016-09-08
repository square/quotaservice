export function formatDate(date) {
  if (!date)
    return ''

  let parsedDate = new Date(date * 1000)
  return `
    ${padZero(parsedDate.getUTCHours())}:${padZero(parsedDate.getUTCMinutes())} \
    ${padZero(parsedDate.getUTCMonth())}/${padZero(parsedDate.getUTCDay())}/${padZero(parsedDate.getUTCFullYear())} \
    UTC
  `
}

function padZero(i) {
  if (i < 10) {
    i = '0' + i;
  }

  return i;
}
