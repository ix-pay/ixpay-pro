// Date类型扩展
declare interface Date {
  /**
   * 将日期转换为指定格式的字符串
   * @param fmt 格式字符串，例如："yyyy-MM-dd hh:mm:ss.S"
   * @returns 格式化后的日期字符串
   */
  Format(fmt: string): string
}
