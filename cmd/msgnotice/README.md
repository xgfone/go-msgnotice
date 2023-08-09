## Message Notice Binary
A simple message notice binary, supporting to send the message notice by HTTP.


### Build
```bash
$ go build
# OR
$ make
```


### API
```http
POST http://localhost/msgnotice
Content-Type: application/json

{
    "Title": "string|required",
    "Content": "string|required",
    "Channel": "string|required",
    "Receiver": "string|required",
    "Timeout": 0, // Optional, Unit: ms
    "Metadata": { // Optional
        "k1": "v1",
        "k2": 123
    }
}


### Success
HTTP/1.1 200 OK
Date: Wed, 09 Aug 2023 10:08:42 GMT
Content-Length: 0


### Failure
HTTP/1.1 400 OK
Date: Wed, 09 Aug 2023 10:08:42 GMT
Content-Type: text/plain
Content-Length: 123

xxxxxxxxxxxxxxxxxxx......
```


### Customization

#### Customize the program initialization process
```
$ go generate -v ./pkg/appinit
```

#### Customize the drivers
```
$ go generate -v ./pkg/drivers
```

#### Customize the middlewares
```
$ go generate -v ./pkg/middlewares
```
