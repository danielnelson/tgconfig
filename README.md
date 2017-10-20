This is an **incomplete** prototype for modifying how influxdata/telegraf
loads its configuration file.

It provides a new type of plugin, a Loader, which allows plugin based loading
and reloading from any data source.

It changes the way plugins are loaded so that they can create a type struct
that is passed to a `New` function to build the plugin.  This should allow one
time initialization and better error reporting.

It also improves the parser and serializer plugins to ease creation of this
type of plugin.

**Running**
```
go run cmd/telegraf/main.go telegraf.conf
```
