// 期限切れ判定はバックエンドのバッチ(日付単位、期日当日は期限内)と基準を揃える。
// due_date は API から "YYYY-MM-DD" 形式で返るため、Date パース(UTC 0時解釈で
// タイムゾーンずれが生じる)を避け、ローカル日付の文字列同士で比較する。
export function isPastDue(dueDate: string): boolean {
  const now = new Date()
  const today = [
    now.getFullYear(),
    String(now.getMonth() + 1).padStart(2, '0'),
    String(now.getDate()).padStart(2, '0'),
  ].join('-')
  return dueDate < today
}
