# Reliable BitTorrent

This file documents any design changes we made to the original BitTorrent.

## Original BitTorrent

### Terminology
 - `torrent`: a structure that corresponds to a file. It contains the `tracker` information and file content hash information. Sometimes a magnet link is used instead, containing similar information. Since `DHT` is not within our scope, it is assumed that it always contains the `tracker` information.
 - `swarm`: a network of nodes in the context of a particular `torrent`.
 - `peer`: a member of a `swarm`. Also sometimes called `client`.
 - `seeder`: a `peer` which has the complete file resource.
 - `leecher`: a `peer` which does not have the complete file resource, and requires uploads from other `peer`s to complete.
 - `tracker`: a coordinator-of-sort in a `swarm`. It helps `peer`s find each other.

 ### Communication
 In original BitTorrent, there are a number of different conversations going on. We will be discussing them in a relatively sequential order:

 #### announce request

 A node accesses a `torrent`, which contains information about who the `tracker` is, as well as the `infohash` of the file. It then send an `announce` message to that `tracker` in an HTTP GET request about its status. Some of the notable parameters in our interests are:
 - `peer_id`: a unique identifier of a `peer`, which the `tracker` uses.
 - `info_hash`: from `torrent`. Works as a unique identifier for the file.
 - `port`: the port the `peer` wishes to use for the communication. Note that this does not have to be the same `port` the `peer` uses to send this `announce` request - it is just for the `peer`-to-`peer` communication.
 - `left`: how many more bytes does it still need to download to have the complete file. This usually starts out as 0 for a `seeder` or the complete file size for a `leecher`.
 - `downloaded`: how many bytes has it downloaded from other `peer`s.
 - `uploaded`: how many bytes has it uploaded to other `peer`s. It is culmulative for each `peer` it uploads to.
 - `num_want`: the maximum number of `peer`s does it wants the `tracker` to provide. Note that the actual list returned by the `tracker` might have a smaller size than this.
 - `event`: an optional parameter that indicates the event, one of `started` (for joining the `swarm`), `completed` (for marking it has the complete file), or `stopped` (for exiting the `swarm`).

 Note that `announce` after the node becomes a `peer` in the swarm still happens periodically (period given in the [announce response](#announce-response)) - this serves as an indicator to the `tracker` that the `peer` is still participating, similar to a periodic `HeartBeat` from some server implementations to express liveness. When it wishes to exit the swarm, it can simply not `announce` so the `tracker` will drop it after a certain period, or `announce` with `event = stopped` for an immediate removal from the records in `tracker`.

#### announce response

Once the `tracker` receives the `announce` request, assuming it is strictly following the rules, then if `announce` indicates `event = stopped` (meaning it is leaving the swarm), respond with an empty HTTP OK and remove it from the tracker's known peers.

Otherwise, the tracker adds or updates its records based on its information given by `peer`, and responds with an HTTP OK with these notable parameters:
 - `interval`: how long until the next `announce` the `peer` should send out.
 - `complete`: the number of `seeders` in the `swarm`.
 - `incomplete`: the number of `leechers` in the `swarm`.
 - `peers`: a collection of `peer`s, each with `peer_id`, `ip` and `port`.

With these information, the receiver `peer` has a list of `peer`s it can try to establish connection to, whether it is to upload to or download from.

#### scrape request
Though not necessarily supported by all `tracker` implementations, a `peer` can send a `scrape` HTTP request to the `tracker` with only the `info_hash` as the sole parameter. This is to ask for status of the `swarm` corresponding to the file without requesting the `peer` lists, which is way more expensive.

#### scrape response
If a `tracker` handles `scrape`, it will simply responds with an HTTP OK with parameters `complete` and `incomplete`, a subset of that of an [announce response](#announce-response).

It is worth noting that in cases where `scrape` is handled, a `peer` will often `scrape` in between `announce`s (especially when there are "bad actors"). If there are too many `scrape`s, the `tracker` will be less performant. Therefore the `tracker` might simply set a limit on the number of `scrape`s it can handle and not respond when the limit is reached.


## ReliableBT

### Terminology
-  `baseline provider`: *A new entity in our modification. It is a consistent node that is not expected to fail easily, can handle a large upload load, and is acting as a special `peer` in a `swarm` to provide service as "minimum guarantee".
