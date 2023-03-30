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
If a `tracker` handles `scrape`, it will simply respond with an HTTP OK with parameters `complete` and `incomplete`, a subset of that of an [announce response](#announce-response).

It is worth noting that in cases where `scrape` is handled, a `peer` will often `scrape` in between `announce`s (especially when there are "bad actors"). If there are too many `scrape`s, the `tracker` will be less performant. Therefore the `tracker` might simply set a limit on the number of `scrape`s it can handle and not respond when the limit is reached.


## ReliableBT

### Terminology
-  `baseline provider`: A new entity in our modification. It is a consistent node that is not expected to fail easily, can handle a large upload load, and is acting similar to a special `peer` in a `swarm` to provide service as "minimum guarantee". For clarification, whenever a "regular `peer`" is mentioned, it is referring to a `peer` that is not the `baseline provider`. Note that a `swarm` can have 0 (just as the original BitTorrent), 1 or multiple of `baseline provider`s.

### Communication

Below are some modified/new communication introduced in ReliableBT.

#### announce request (for baseline provider)

Though the `announce` coming from a regular `peer` will be the same, a `baseline provider` will also `announce` itself to the `tracker` periodically. The reason why we don't have the `baseline provider` info directly in `torrent` is so that:
 - for any file, the `torrent` should stay the same. If there is ever a change to a `baseline provider` (for example a change in the number of these, or simply changing from one IP/host to another), this would fail to work.
 - sometimes a `baseline provider` might be down (e.g. for maintainence). The `swarm` will be more performant if each `peer` has to talk with the `tracker` to know whether there are `baseline provider`s currently usable, rather than trying to contact the `baseline provider`s given in the `torrent` right away and fail.

This, of course still means that the `tracker` will be the single point-of-failure in this system, because a `peer` will never be able to learn about `baseline provider` without contacting the `tracker`. Our model assumes the reliablity of the `tracker` and any version of BitTorrent without it (for example, with a DHT) is outside our scope.

Now regarding the format of this `announce` call, it will be similar to that of the basic [announce request](#announce-request). Again as an HTTP Get request, it will have these following changes:
 - `baselineProvider`: a new parameter that is optional. If given and has a `true` value, the sender is assumed to be a `baseline provider` for the given `infoHash`.
 - `downlodaded`: no longer a required parameter when `baselineProvider = true`.
 - `uploaded`: no longer a required parameter when `baselineProvider = true`.
 - `left`: no longer a required parameter when `baselineProvider = true`, as the file is assumed to be already complete.
 - `event`: only has 2 possible values when `baselineProvider = true`, `started` and `stopped`. `completed` is redundant similar to `left`.

#### announce response (for baseline provider)

In response to this [announce request](#announce-request-for-baseline-provider) for `baseline provider`, there is also minor modification to the action of the `tracker`.

If it sees `baselineProvider = true`, `tracker` will store the information separated from the regular `peer`s. It will send the exact same [announce response](#announce-response).

This simplfied model comes with an issue that anyone can pose as a `baseline provider` just by setting a flag, especially from malicious actors. The current implementation simply has `tracker` keeping records of trusted IP values to verify the truthfulness of `baselineProvider = true`. This is but a rough implementation, but as `tracker` needs to potentially handle requests from `peer`s of multiple `swarm`s in a short time frame, sticking with the one-time HTTP request/response is what our implementation sticks to currently.

Anyone who poses as a `baseline provider` but turns out to be a fake will have its message dropped and no response will be sent.

#### announce response (for regular peer)

Now with the addition of `baseline provider`, regular `peer` will get information about it as well when it `announce`s. Notably, below is the change to the parameters compared to the original [announce response](#announce-response):
 - `peers`: this now only applies to regular `peer`s that are not the `baseline provider`s.
 - `baselineProvider`: this is an optional parameter, representing a `baseline provider` from its known set of them. Similar to `peers`, this includes its `peer_id`, `ip` and `port`.

For which `baseline provider` is chosen to be promoted to a `peer` when multiple are known to the `tracker`, this will be described in the [tracker details](#tracker-details) section below.

For how a "bad actor" is identified and how to respond in this case, this will be described in the [bad actors](#bad-actors) section below.

### Baseline Provider Selection

### Bad Actors
