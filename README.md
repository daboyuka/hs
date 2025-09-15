# hs

`hs` is a tool for making batch, data-driven HTTP requests.

```
# single requests
hs GET '//example.com'
hs POST '//example.com' 'some data'

# request per line in file: replace end of URL with line
hs GET '//example.com/person/${.}' <people.txt

# request per JSON value in file: replace end of URL with "name" field, body with "address" field
hs -i json PUT '//example.com/person/${.name}/address' '${.address}' <people.json

# request per JSON value, but up to 64 requests in parallel
hs -i json -P 64 POST '//example.com/upload/${.id}' '${.data}' <uploads.json
```

## Commands and Flags

This is a summary; run with `-h` for full help.
```
COMMANDS:
hs <method> [cflags] [bflags] [rflags] <url> [<body>]
  build and run HTTP requests, print responses on stdout
  
hs build [cflags] [bflags] [-X <method>] <url> [<body>]
  build HTTP requests, print requests on stdout (for use with "run" later)

hs run [cflags] [rflags]
  run HTTP requests read from stdin, print responses on stdout

hs [<command>] -h
  print full help [for <command>]

ARGS:
method: the HTTP method (default GET)
url   : the URL: may omit scheme (default "https:") or hostname (default $HOST) (templated)
body  : the request payload, for valid methods (templated, default empty)
```

```
FLAGS:
cflags (common flags):
  -i, --infmt: input format: auto, null, raw, lines, json, [raw]csv, [raw]tsv
  (default 'auto': 'null' if tty stdin, otherwise autodetect as 'json' or fallback to 'lines')

bflags (build flags):
  -H hdr             : add an HTTP request header, format "key: val" (templated , may be repeated)
  -L, --loadjson arg : load a JSON file as a lookup table; argument has syntax "filename,varname,keyexpr" 
                       filename = file to load, varname = variable to load into (as an object record),
                       keyexpr = expression to extract the key for each loaded value, to store it as an entry in varname
rflags (run flags):
  -b name=value   : add a single cookie with name/value (may be repeated)
  -b cookiefile   : add a cookiejar file, curl/Netscape format (may be repeated)
  -F, --fails file: write failure responses (conn. error / non-2xx status) to file, or to
                    stdout if "-" (default "-")
  -o, --outfmt    : response output format: one of body (payload only; default), bodycode (status + payload), 
                    reqresp (request + response), or resp (response only)
  -P pll          : run at most pll requests in parallel (default 1)
  -p mode         : show progress bar; mode one of "auto" (if stdout redirected to file; default), "true", "false"
  -r retries      : retry failed requests (conn. error / non-2xx status) up to retries times
```

## Data-driven HTTP

### Records

Commands iterate over a stream of "records", or values with JSON datatypes:
null, boolean, number (float), string, array, object:
* `build`: any record &rarr; construct an HTTP request &rarr; `request` record   
* `run`: request record &rarr; run the HTTP request &rarr; `reqresp` record (or other, based on -o flag)
* `<method>`: any record &rarr; chain of `build` + `run` &rarr; `reqresp` record

These commands produce/expect `request`, `response`, or `reqresp` records with these schemas: 
```
request  = {"method":"...", "url":"...", "headers":<headers>, "body":"..."}
response = {"status":123, "headers":<headers>, "body":"..."}
reqresp = {"response":<response>, ...other keys as request...}
headers  = {"hkey1":"hval1" OR ["hval1",...], ...}
```

For a request where an HTTP protocol or network error occurred instead of a server response,
`response` instead has this schema:
```
response = {"error":"the error here"}
```

If retries occurred (see `-r` flag), `response` will have an added `retries` key, an array of
objects with `response` schema for responses prior to the final response.

### Templating and Expressions

**Templates:** some arguments (url, body, headers) are "templated", which may contain escapes like
`${<expr>}` (expression syntax below), to be evaluated/substituted on each input record, allowing
data-driven behavior.

Examples:
```
hs build -X PUT '//${.customer}.acme.com/api/item/${.item.k}' '${.item.v}'
   {"customer": "foocorp", "item": {"k": "foo", "v": "bar"}}
   {"customer": "barcorp", "item": {"k": "baz", "v": [1,2,3]}}
=> {"method": "PUT", "url": "https://foocorp.acme.com/api/item/foo", "body": "bar"}
   {"method": "PUT", "url": "https://barcorp.acme.com/api/item/baz", "body": "[1,2,3]"}
```

**Expressions:** evaluated in templates to compute replacement text. Syntax:
```
.foo        lookup field "foo" in current record (object)
foo         lookup variable "foo" (lowercase)
FOO         lookup global/config variable "FOO" (uppercase)
.[<expr>]   lookup that field/array-idx <expr> (evaluated) in record (object/array)

"string"    string literal
            special: escape \(<expr>) templates <expr> into the string
123         numeric literal (integer only)
```

Examples:
```
.foo[123].bar["baz"][321]
FOO[123]
.table[.key]
```

## Config and Globals

At startup, HScript loads global variables, with names uppercased, from list of places below.
These variables may be used in templates/expressions, and some have special effects (see below).

Load order (later overrides earlier):
* YAML file `$HOME/.hs`
* YAML file `./.hs` (i.e. in current working directory)
* Environment variables with prefix `HS_`, after dropping prefix

For YAML files: top-level keys (uppercased) become global variables. Example:
```
# This file binds two globals:
#   MYVAR to number 123
#   MYTABLE to object {"foo":"bar","baz",321}

myvar: 123
mytable:
  foo: bar
  baz: 321
```

For environment variables: values are always interpreted as strings.

### Special Globals
These global variables have additional, special behavior when bound:
```
HOST               : default hostname (when omitted from url) used in build command (string)
HOST_ALIASES       : a mapping XXX->YYY that causes request building to replace hostname '@XXX' (note the @) with YYY
COOKIES            : extra cookies (if "name=value") or cookiejar files (otherwise) to use (string or array-of-strings)  
BROWSER_LOADERS    : browser(s) to autoload cookies from, or "all" for all supported
                     see section below (string or array-of-strings)
COOKIE_HOST_ALIASES: a mapping XXX->YYY that causes requests for hostname XXX to also match cookies for hostname YYY
                     (in addition to matching XXX as normal).
                     If YYY has a scheme (e.g. http:// or https://), cookies are matched using that scheme rather than
                     that of the original request.
                     Example: XXX -> https://YYY will allow requests to http://XXX to use Secure-only cookies from YYY.  
```

Example:
```
host: example.com
host_aliases:
  coolhost: example.com  # //@coolhost/... becomes //example.com/... 
cookies:
  - cookiename=cookieval
  - cookiejarfile.txt
browser_loaders:
  - chrome
  - firefox
  - safari
cookie_host_aliases:
  example.com: foobar.com  # //example.com/... will match both normal cookies _and_ cookies as if it were //foobar.com/... 
```

## HTTP Engine

### Ctrl+C (interrupt)

If `hs` receives SIGINT (Ctrl+C) while running HTTP requests, it will finish any in-flight requests normally, but
subsequent requests will return immediately with response `{"error":"request not sent"}`. If SIGINT is sent again,
`hs` will terminate in-flight requests (with typical response `{"error":"...: context canceled"}`, but not guaranteed).

### Content-Type
If a request has a body _and_ no `Content-Type` header is given, `hs` will try to autodetect and set
the header. On the _first_ request (that meets these criteria), it uses heuristics on (up to) 512 body
bytes to autodetect; later requests reuse the autodetected type. These `Content-Type`s are understood:
```
application/json
application/x-www-form-urlencoded
```

### Cookies
Cookies are always loaded from several places, listed below.
However, the only cookies sent with a request are those "applicable" to the request's hostname, as
per [RFC 6265](https://datatracker.ietf.org/doc/html/rfc6265#section-5.4)
(exception: "literal" cookies, `name=value` format added by `-b` flags or config, are always sent).

"Cookiejar files" must be in curl/Netscape syntax.

Cookie load locations:
* `-b` flags: may include `name=value` literal cookies and cookiejar filenames
* `.hscookie` cookiejar files: looks in home directory and/or current directory (curl/Netscape syntax)
* `COOKIES` config var: may include `name=value` literal cookies and/or cookiejar filenames
* "Browser loaders": loads cookie store from browsers as controlled by `BROWSER_LOADERS` config var (see below)

### Browser loaders
HTTPScript can load cookies directly from browsers installed on your system. By default, this is disabled, but
can be enabled by populating the `BROWSER_LOADERS` config var a (list of) browser name(s).
Run `hs help` to see list of supported browsers.

When enabled, all cookies are loaded by default from configured browser stores. To restrict this, also
set the `BROWSER_LOADER_PREFIXES` config var to a (list of) cookie name prefix(es) to load.

## Design notes

### Goals

1. Economy (simple cases are simple; defaults for everything)
2. Streaming (continuous output)
3. stdin/stdout-oriented
4. Repeatable (scriptable, independent build/run)
5. Retriable (easy to rerun failed/all requests, immediately or later)
6. Extensible (domain-specific tooling)

### Limitations
The 'string' datatype stores character data, not bytes. The internal encoding is UTF-8. HScript cannot
handle raw binary data at the moment.

### Scripting (WIP)

In the future, HScript will support script files with sequences of commands, including
`build`, `run`, and `<method>`, but also with record-handling commands such as parsing,
unwinding/collecting, etc.

**Command:** a top-level action, e.g. send an HTTP request. A command is repeated for
each input "record", and usually produces output. An `hs` terminal one-liner (vs.
using an `hs` script) is a single command.
