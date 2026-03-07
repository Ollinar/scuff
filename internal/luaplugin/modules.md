


## json

### decode

```lua
decode(json_string) -> any, error
```
decode takes a string and returns the parsed value and an error string if any is encountered.

### encode

```lua
encode(any) -> json_string, error
```
decode takes a value and returns the encoded string and an error string if any is encountered.

## http

### request

```lua
request(http_method,url,(optional)opts) -> response_body, status_code, response_header, error
```
the opt is table with 2 available fields, headers and body. headers should be a key value pair string, while body is string.
#### example

```lua
local body, status_code, headers, err = http.request("POST", "https://www.google.com", {
        headers = {
            val1="a",
            val2="b",
        },
		body = "{json_payload}",
	})
```

## util

### htmlUnescape

```lua
htmlUnescape(html_escaped_string) -> unescaped_string
```

### htmlEscape

```lua
htmlEscape(some_string) -> html_escaped_string
```
