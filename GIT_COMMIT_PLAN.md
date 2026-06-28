# Git コミット計画

| # | コミットメッセージ | 対象ファイル |
|---|---|---|
| 1 | `chore: add .gitignore` | `.gitignore` |
| 2 | `chore: add CI/CD workflows` | `.github/` |
| 3 | `chore: add docker-compose` | `docker-compose.yml` |
| 4 | `feat: implement backend` | `backend/` |
| 5 | `feat: implement frontend` | `frontend/` |
| 6 | `feat: implement lambda` | `lambda/` |
| 7 | `chore: add Claude agent configs` | `.agents/`, `.claude/`, `skills/`, `skills-lock.json` |

## 注意事項

- `.gitignore` を最初にコミットして秘密鍵・`.env` の混入を防ぐ
- `backend/todo-8a6ad-firebase-adminsdk-fbsvc-cc3c7e2756.json` は除外済み
- `frontend/.env` は除外済み → GitHub Variables で代替（`frontend/.env.example` 参照）
