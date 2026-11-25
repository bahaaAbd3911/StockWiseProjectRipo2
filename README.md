# StockWise

An online-shop platform: one codebase and one deployment run many independent
stores, each with its own address (or custom domain), products, customers,
orders, look and admin - in English and Arabic.

- **Read first:** [DOCUMENTATION.md](DOCUMENTATION.md) - what everything does and why it was built this way.
- Backend reference: [backend/docs/GUIDE.md](backend/docs/GUIDE.md)
- Server checklist: [backend/docs/production.md](backend/docs/production.md)

```powershell
cd backend; go run ./cmd/migrate up; go run ./cmd/api     # API on :8080
cd Front-End; npm install; npm run dev                     # shop on http://localhost:5173
```
