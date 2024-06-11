# Chat Server
 Real time chat room exercise.
 
## Why
Wanted to try out [Templ](https://templ.guide) and [HTMX](https://htmx.org) and practice go channels.

## Build and Run
Build the Templ files & run the server
```bash
./run.sh
# Then open http://localhost:8888
```

## Watch
Watch for changes in both Templ files and go files and re-load the app on changes.  
_We cannot do live reloading because proxies work bad with SSE._
```bash
./watch.sh
# Then open http://localhost:8888
```