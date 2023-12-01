In the exec, container id needs to seleted

```javascript
var ws = new WebSocket('ws://localhost:8080/ws');
```


```javascript
ws.onmessage = function(event) {
    var data = JSON.parse(event.data);
    console.log('Server: ' + event.data);
};
```


The payload format needs to match this-

##
```javascript
ws.send(JSON.stringify({
    operation: "exec",
    containerID: "your-container-id",
    command: ["apk", "update"]
}));
```

if we do not provide the container id, it response "No containers available to execute command."

### Screenshots

[]