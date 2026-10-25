/**
 * 部署脚本：打包完成后将 dist 目录全量复制到 Nginx 静态目录
 * 流程：先删除目标目录，再整体复制（全量替换）
 */
import fs from 'node:fs'
import path from 'node:path'

const sourceDir = path.resolve(process.cwd(), 'dist')
const targetDir = 'D:\\1_Software\\Docker\\Nginx-1.30.0\\html\\wxapp'

// 校验构建产物存在
if (!fs.existsSync(sourceDir)) {
  console.error(`[部署失败] 未找到构建产物目录：${sourceDir}，请先执行打包`)
  process.exit(1)
}

try {
  // 先删除旧的目标目录（全量替换）
  if (fs.existsSync(targetDir)) {
    fs.rmSync(targetDir, { recursive: true, force: true })
    console.log(`[部署] 已删除旧目录：${targetDir}`)
  }

  // 复制构建产物到目标目录
  fs.cpSync(sourceDir, targetDir, { recursive: true })
  console.log(`[部署成功] 已将 ${sourceDir} 全量复制到 ${targetDir}`)
} catch (err) {
  console.error(`[部署失败] ${err.message}`)
  process.exit(1)
}
