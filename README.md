### modlog

`modlog` produces a verbose log of your dependency version bumps

### Installation 
```
go get github.com/dprotaso/modlog
```

### Usage

```
$ modlog knative.dev/pkg origin/release-0.15 origin/release-0.16
bumping go.opencensus.io d835ff8...5fa069b:
  > 5fa069b Initialize View Start Time During View Registration (#1215)
  > 1901b56 Allow custom view.Meters to export metrics for other Resources (#1212)
  > 785d899 Delete views from measure ref when unregistering (#1211)
  > cd9ae5c Remove call to time.Now() on worker thread when handling record reqs (#1210)
  > 46dfec7 Reduce allocations (#1204)
  > d3cf45e Safely reject invalid-length span and trace ids (#1206)
  > 84d38db Allow creating additional View universes. (#1196)
  > a7631f6 replace gofmt with goimports (#1197)
...
```

