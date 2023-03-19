# Reliable BitTorrent Validation
This is the validation component for the Reliable BitTorrent group project (UBC CPSC 416 2022W2). The client (peer) repository can be found [here](https://github.com/Maxwell-Yang-2001/reliableBT-tracker), and the tracker repository can be found [here](https://github.com/kaiyyang/cpsc416_GroupProject_ReliableBT).

## Prerequisite
1.
    The setup is only guaranteed to work on macOS with M1 or newer chips (the ones used by all our development members). It should work on other architectures/OSs with potential minor changes, but due to development limitation, no guarantee can be given.

2.
    Please make sure your Golang is installed with version >= 1.20. You can verify it through entering the following to your command line:
    ```sh
    go version
    ```

3.
    This repository should be cloned as a silbing to the ReliableBT repository in your local file system, like the following:
    ```
    parent_directory
    ├- cpsc416_GroupProject_ReliableBT
    └- reliableBT-validation
    ```
    If for some reason this is not possible on your machine, run the followings in the root of this repository:
    ```sh
    go mod edit -replace github.com/anacrolix/torrent=<PATH TO THE RELIABLEBT REPOSITORY>
    go mod tidy
    ```

4.
    Please make sure both this repository and the ReliableBT repository are up-to-date - you can fetch from upstream if necessary.

## Setup Validation
To confirm your setup is ready, below are the steps to run a simple test to verify the connection between the 2 repositories:

1.
    In the root of your ReliableBT repository, switch to the `validation-setup` branch (that simply adds a dummy function):
    ```
    git checkout validation-setup
    ```
2.
    In the root of this current repository, run the setup tests (check whether the dummy function is callable):
    ```
    cd tests
    go test -v
    ```
    The test should pass with "Setup Successfully" logged. You can modify the dummy function in `setup.go` and fail the test to verify the consistency further.
3.
    In the root of your ReliableBT repository, switch back to master:
    ```
    git checkout master
    ```
And your setup should be good to go: any import reference to `github.com/anacrolix/torrent` in this repository should be pointing to the local clone of your ReliableBT repository.
