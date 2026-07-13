# Build Baseline

## Environment

- Go: `go version go1.26.4 windows/amd64`
- Node: `v24.13.0`
- npm: `11.6.2`
- Wails: `v2.11.0`

## Commands

```powershell
npm --prefix frontend run build
go test ./... -run '^$' -count=0
wails build
```

## Expected Result

- Frontend production build succeeds.
- Go packages compile without running integration tests.
- Wails creates `build/bin/go-stock.exe`.

## Known Warnings

- Vite may warn about large chunks.
- Wails binding generation may print unresolved helper types such as `time.Time` or `gorm.DeletedAt`; these warnings do not currently block the build.
