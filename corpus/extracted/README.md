# Extracted vectors

This region of the corpus contains vectors that have been extracted from live
networks using the [`tvx` tool](https://github.com/filecoin-project/lotus/tree/master/cmd/tvx).

It is further classified in **batches**.

## Batches

Each subdirectory maps to a batch.

Once committed, a batch is immutable (barring structural schema
upgrades/migrations).

Batches are identified by four-digit zero-left-padded sequence numbers,
starting from 0001.

To enhance searchability, batches can be labelled by appending a kebab-cased
suffix. The label can capture the temporality, functional scope, or technical
nature of the enclosed vectors, or otherwise be a meaningful description that
facilitates discovery and categorisation.

Labels may be (1) repeated, but sequence numbers cannot; (2) modified at a later
time, such as to disambiguate a future, incoming batch.

In a nutshell:
- batch labels are mutable, optional, and non-unique.
- batch sequence numbers are immutable, compulsory, and unique.

Example:

```
corpus/
  |__ extracted/
        |__ 0001-initial-vectors/
        |__ 0002-payment-channels/
        |__ 0003-network-v1-modified-actors/
        |__ 0004-network-v1-modified-actors/
        |__ ...
```

## Batch contents

Each batch contains one or many test vectors.

Batches can be further broken down into arbitrary subdirectory structures, for
finer vector classification.

For example, the `0001-initial-vectors` batch has this structure:

```
0001-initial-vectors/
  |__ [receiver actor code] (replacing slashes with underscores)
        |__ [method name]
              |_ [exit code; string representation if available]
                   |__ vector1.json
                   |__ vector2.json
                   |__ ...
```

It is advised for batches to contain a `README.md` file that outlines, at least:

- (Approximate) extraction timestamp.
- Description of the batch contents/scope.
- Message selection technique: heuristics used to identify messages, e.g.
  "latest 10 messages with a unique (actor_code,method_num,exit_code) tuple, up
  to chain epoch N".
- To maximise traceability and reproduceability, scripts and technical
  assets are welcome, e.g. Python scripts against the JSON-RPC API, or SQL
  queries against chainwatch/sentinel databases.

Assets other than test vectors MUST NOT carry `.json` extensions, to avoid being
confounded with vectors.
