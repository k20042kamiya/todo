# バックエンド実装ガイドライン

## バックエンドコードレビュー時（必須）

`backend/` 配下の `.go` ファイルをレビュー・実装・編集した後は、
必ず以下のskillを使ってレビューを行い、問題があれば即座に修正してから完了を報告すること。

```
use backend-review skill
```

- Echo middleware / context handling は Context7 で最新ドキュメントを確認する
- GORM transactions / preload / N+1 は Context7 で最新ドキュメントを確認する
- Clean Architecture の依存方向（handler → usecase → domain）を確認する
- `context.Context` がリクエストスコープから全レイヤーに伝播しているか確認する

### レビューフロー

1. 実装完了
2. `backend-review` skillでチェック（必要に応じてContext7を使用）
3. 問題があれば修正
4. 「backend-review skillチェック: 問題なし」を確認してから完了報告

---

# Vue実装ガイドライン

## Vueコード実装後のセルフレビュー（必須）

`frontend/` 配下の `.vue` または `.ts` / `.tsx` ファイルを実装・編集した後は、
必ず以下のskillを使ってセルフレビューを行い、問題があれば即座に修正してから完了を報告すること。

### チェックするskill

```
use vue-best-practices skill
```
- `<script setup lang="ts">` + Composition API を使っているか
- `ref` / `shallowRef` / `computed` の使い分けが正しいか
- `v-for` に `:key` があるか、`v-if` と `v-for` を同一要素に併用していないか
- コンポーネント分割・composable 設計が適切か

```
use vue-router-best-practices skill
```
- ルートパラメータの変化を `watch(route.params, ...)` で監視しているか
- ナビゲーションガードの使い方が正しいか
- ルートとコンポーネントのライフサイクルが適切に扱われているか

```
use vue-debug-guides skill
```
- ランタイムエラーや警告の原因になりやすいパターンがないか
- 非同期処理のエラーハンドリングが適切か

### レビューフロー

1. 実装完了
2. 上記3つのskillでチェック
3. 問題があれば修正
4. 「Vue skillチェック: 問題なし」を確認してから完了報告
