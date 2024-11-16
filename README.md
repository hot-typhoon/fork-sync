# fork-sync

一个简单的Vercel应用  
使用webhook调用Vercel Function，将fork仓库和上游同步

```txt
POST https://xxx.vercel.app/api/sync?owner=xxx&repo=xxx&branch=xxx&pat=xxx&only=xxx&shallow=xxx
```

```golang
type QueryParams struct {
    Owner   string
    Repo    string
    Branch  string
    Pat     string
    Only    string `sync:"none"`  // optional, can be "git" or "api"
    Shallow string `sync:"false"` // optional, can be "true" or "false"
}
```
