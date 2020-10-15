# Batch 0002: Init Actor

> first extracted around Wed Sep 30 14:00:00 UTC 2020 (tvx version: eb6191d0ffd01a7cf7f8544a31acf307b1799fb2)
>
> reextracted around Wed Oct 14 16:30:00 UTC 2020 (tvx version: https://github.com/filecoin-project/lotus/pull/4393)

This is a selection of messages to the singleton init actor, extracted from
heights 47000-50000 of the Space Race chain (which later transitioned to
Ignition), using the `tvx` extraction tool.

## Message selection

Init actor messages selected by querying a Chainwatch DB instance that was
synced up to chain height ~57000.

## SQL query (against chainwatch)

Resulted in [selection.csv](./selection.csv).

```sql
with uniq_msgs as (
    select msgs.cid                     as message_cid,
           actors.code                  as receiver_code,
           msgs.method                  as method_num,
           receipts.exit                as exit_code,
           blocks.height                as height,
           blocks.cid                   as block_cid,
           row_number()
           over (partition by msgs.cid) as uniq_rn
    from public.messages as msgs -- join messages, with their blocks, their actor types, and receipts.
             join public.block_messages as block_msgs on msgs.cid = block_msgs.message
             join public.blocks as blocks on blocks.cid = block_msgs.block
             join public.actors as actors
                  on msgs.to = actors.id and actors.stateroot = blocks.parentstateroot -- this is not precise, but actor types are immutable, so it'll suffice
             join public.receipts as receipts on msgs.cid = receipts.msg
    where actors.code = 'fil/1/init'
    order by height desc
),
     group_by_type
         as -- take the previous input, and assign row numbers based on message_cid; we'll only retain unique messages.
         (select uniq_msgs.*,
                 row_number() over (partition by receiver_code, method_num, exit_code order by height desc) as group_rn
          from uniq_msgs
          where uniq_rn = 1
          order by height desc
         )
select message_cid, receiver_code, method_num, exit_code, height, block_cid, group_rn as seq
from group_by_type
where group_rn <= 10
order by receiver_code, method_num, exit_code, height desc
;
```