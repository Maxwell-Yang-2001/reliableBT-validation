# Reliable BitTorrent Validation
This is the validation component for the Reliable BitTorrent group project (UBC CPSC 416 2022W2). The client (peer) repository can be found [here](https://github.com/kaiyyang/cpsc416_GroupProject_ReliableBT), and the tracker repository can be found [here](https://github.com/Maxwell-Yang-2001/reliableBT-tracker).

## Prerequisite
1.
    The setup is only guaranteed to work on macOS with M1 or newer chips (the ones used by all our development members). It should work on other architectures/OSs with potential minor changes, but due to development limitation, no guarantee can be given.

2.
    Please make sure your Golang is installed with version >= 1.20. You can verify it through entering the following to your command line:
    ```sh
    go version
    ```

3.
    Please make sure you have at least 65GB of free space on your device, as some tests involve transfer of large files.

4.
    This repository should be cloned as a sibling to the client and tracker repository in your local file system, like the following:
    ```
    parent_directory
    ├- cpsc416_GroupProject_ReliableBT
    ├- reliableBT-tracker
    └- reliableBT-validation

5.
    Please make sure all repositories in the point above are up-to-date - you can fetch from upstream if necessary.

## Setup
To setup, simply run the test file `./tests/setup_test.go`:
```
go test test/setup_test.go -v
```

It runs some tests to check your local repositories have been set up correctly, as well as generating randomized test files of various sizes under `./resources` (you should see `test_1kb.txt`, `test_1mb.txt`, `test_1gb.txt`, `test_10gb.txt` and `test_20gb_txt` inside this directory).

## FAQ

1.
    Q: I am using VSCode. I ran a test through the UI (instead of terminal) which indicated that the test timed out after 30 seconds. Is this expected?
    
    A: That is indeed expected! Some of our tests are expected to run for a long time, since performance is a criteria we want to measure. Running the test through the UI simply adds a flag to time out with the 30 seconds by default of the Go VSCode extension.
    
    To work around it, go to `settings.json` of your VSCode (you can use the command palette and search for "Preference: Open Settings (UI)"). Then simply find the setting corresponding to the test timeout (it has a setting ID of `go.testTimeout` which you can type in the search bar to filter) and adjust its value, which we recommend to be 300s.