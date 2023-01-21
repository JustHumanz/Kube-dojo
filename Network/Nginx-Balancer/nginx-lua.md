## Nginx-lua

- `wget -O - https://openresty.org/package/pubkey.gpg | sudo apt-key add -`
- `echo "deb http://openresty.org/package/ubuntu $(lsb_release -sc) main" > /etc/apt/sources.list.d/openresty.list`
- `apt update`
- `apt install openresty`
- `systemctl stop openresty.service`
- `mkdir nginx-lua;cd nginx-lua`
- Download nginx.conf
- `openresty -c $(pwd)/nginx.conf -t` # Make sure the test was successful
- `openresty -c $(pwd)/nginx.conf`
- `go build -o simple_web .`
- `for i in {1..3};do (./simple_web 127.0.0.1:111$i web_$i &) ;done`
- `curl -X POST -d '{"action":"add","ip":"127.0.0.1","port":1111}' http://localhost:9091/config`
- `crul localhost` # The output should Hello from web_1
- `curl -X POST -d '{"action":"add","ip":"127.0.0.1","port":1112}' http://localhost:9091/config`
- `curl localhost` # The output should Hello from web_1 or web_2
