//  字典文字展示方法
export const showDictLabel = (dict, code, keyCode = 'value', valueCode = 'label') => {
  if (!dict) {
    return ''
  }
  const dictMap = {}
  dict.forEach((item) => {
    if (Reflect.has(item, keyCode) && Reflect.has(item, valueCode)) {
      dictMap[item[keyCode]] = item[valueCode]
    }
  })
  return Reflect.has(dictMap, code) ? dictMap[code] : ''
}
