## Instructions
Below are temporary instructions to get the trackers working. A lot of things require refining, however it is the first priority to make sure everyone at least has the setup ready.

1.
    Please clone the tracker repo as siblings to this repo and the core repo: https://github.com/crimist/trakx.
2.
    Make sure you don't have any process occupying port 1337 (I have not yet figured out how to customize trakx ports properly, so this is a temporary solution) by running:
    ```sh
    lsof -i:1337
    ```
    If nothing gets printed, port 1337 is available. Otherwise, you need to kill the process on this port (the pid information will be available from this command, then just `sudo kill` it).
3.
    Inside the root of trakx repo, run:
    ```sh
    go build
    ./trakx start
    ./trakx status
    ```
    `start` simply kicks off the tracker, while `status` helps you identify whether the service is properly working (you should see 3 green checkmarks).
4.
    Now that tracker is properly running, inside the validation repo, do (you might have to `go mod tidy` first):
    ```sh
    go test basic_test.go -v
    ```
    This file is running 2 tests, one being a direct seeder->leecher transfer with leecher knowing who is a seeder from the get-go, and the other test including a tracker for peer-discovery between them.
    
    If things are set up correctly, both tests inside this file should pass, and you should see a hello.txt (used by test 1) and C#.pdf (used by test 2) being transferred to under tests/leecher from tests/seeder.
