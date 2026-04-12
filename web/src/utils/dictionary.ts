// 字典文字展示方法
export const showDictLabel = <T extends object>(
  dict: T[] | undefined,
  code: string | number,
  keyCode: keyof T = 'value' as keyof T,
  valueCode: keyof T = 'label' as keyof T,
): string => {
  if (!dict) {
    return ''
  }
  const dictMap: Record<string | number, string> = {}
  dict.forEach((item) => {
    if (Reflect.has(item, keyCode) && Reflect.has(item, valueCode)) {
      dictMap[item[keyCode] as string | number] = item[valueCode] as string
    }
  })
  return Reflect.has(dictMap, code) ? dictMap[code] : ''
}
