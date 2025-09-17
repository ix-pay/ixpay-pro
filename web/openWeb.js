import child_process from 'child_process'

// 从环境变量中获取端口，如果没有则使用默认值
const port = process.env.VITE_CLI_PORT || '8585'
var url = `http://localhost:${port}`
let cmd = ''
switch (process.platform) {
  case 'win32':
    cmd = 'start'
    console.table({
      platform: process.platform,
      cmd,
      url
    })
    child_process.exec(cmd + ' ' + url)
    break

  case 'darwin':
    cmd = 'open'
    console.table({
      platform: process.platform,
      cmd,
      url
    })
    child_process.exec(cmd + ' ' + url)
    break
}
