# Batch 0001: Initial extraction

> first extracted around Wed Sep 30 11:00:00 UTC 2020 (tvx version: eb6191d0ffd01a7cf7f8544a31acf307b1799fb2)
>
> reextracted around Wed Oct 14 16:00:00 UTC 2020 (tvx version: https://github.com/filecoin-project/lotus/pull/4393)

This is a selection of 201 messages extracted from heights 65000-69000 of
the Space Race chain (which later transitioned to Ignition), using the `tvx`
extraction tool.

## Message selection

Ten most recent messages a unique combination of
`(actor_code,method_num,exit_code)`, from chain heights 65000-69000.

## Unsuccessful extraction

The following messages were not extracted successfully due to mismatches in
local execution receipt vs. chain receipt. Once issues are fixed, a future batch
may incorporate these vectors. They appear in [selection.csv](./selection.csv),
but do not have corresponding vectors.

```
* message bafy2bzacecmvh7pwndzschh3yvqtck42vyzshe2rfc3ssqcxo55ejpbtiook2: vector generation aborted (receipt mismatch)
* message bafy2bzacedrxvrl2bhkdwriyztxrau2vencnu7bjupgrrhfyaamj4n6lszmxw: vector generation aborted (receipt mismatch)
* message bafy2bzaceb5gw4gesypkat7sapwhkvfge62rqnin6lj24mcvvnmvmrhvb7oru: vector generation aborted (receipt mismatch)
* message bafy2bzacedjsyltdtdapif43qf7wf5xa2uebtw3qbrypmdufanpqnfzegjqsu: vector generation aborted (receipt mismatch)
* message bafy2bzaceccjvw46aw6mwwr4tqrii2fw5qjsjbnz7pwv7itqoep4ztvf4sqhy: vector generation aborted (receipt mismatch)
* message bafy2bzacearefgbehpyizzaxe52pv5lzltzjv4od6v734b36vgep4zkzim7ti: vector generation aborted (receipt mismatch)
* message bafy2bzacedzpgzx4fbcqgkqh3l4edhsjw62rl6rh2j4unfglq7c3rw347h6vu: vector generation aborted (receipt mismatch)
* message bafy2bzacedoegd7reqzpdjvaxjtuqtzuprqbily2o53wmnfputbc52j4k3lhs: vector generation aborted (receipt mismatch)
* message bafy2bzaceddkzinn35g3iro2px5oo2eavjsghispsx24bivlcrxhrfhvwigbq: vector generation aborted (receipt mismatch)
* message bafy2bzacedxsdpaiqov3dddjib4jmp5eoehulh6gpmault2jxod5d523olsz6: vector generation aborted (receipt mismatch)
* message bafy2bzacebpxw3yiaxzy2bako62akig46x3imji7fewszen6fryiz6nymu2b2: message not found; precursors found: 332
* message bafy2bzacebs43sf277jlreduqkx3zaljurt675dvs7rg4vvkst3k3cqtggena: vector generation aborted (receipt mismatch)
* message bafy2bzacebe2kzhdjjtquj5yof32o3oop2ffubhsoqchn42jlezhtvpisqnf6: vector generation aborted (receipt mismatch)
* message bafy2bzaced7k7wygzqianzymmnyeb56flcen33mzurax5ok3jjeondwmg2o7g: vector generation aborted (receipt mismatch)
* message bafy2bzacebj67kfde3yn2gcse6rwuz4ko67fj4ai4xciylcgwuqyd6wvhf6ak: vector generation aborted (receipt mismatch)
* message bafy2bzacec7ceff2l7slqjdsptidd73nuy7kmiqecyqmufhgnalwy4dbvc5tm: vector generation aborted (receipt mismatch)
* message bafy2bzaced2tjkxc77s34hajk4ckylh3zgpkoojypwexrr5eamdwi527lbnha: vector generation aborted (receipt mismatch)
* message bafy2bzacedehjkw77xl7sc2fqh6o3v63phs4z2ijawtujyxxzq4sw6s6ait46: vector generation aborted (receipt mismatch)
* message bafy2bzaceack4i2a73uav3juaxrae2yt3lcztjmd5tg7osttrfamawmfc6lyo: vector generation aborted (receipt mismatch)
* message bafy2bzacebte3egdqgx6qnvsh7ubgnkjxu7hxzcdzfjhuioplb56eeb5zvzls: vector generation aborted (receipt mismatch)
* message bafy2bzacecpb6h5ahvmt5etmcug2zmzkoozk74oypmbejxl4cgbqpf7syyies: vector generation aborted (receipt mismatch)
```

## SQL query (against sentinel/visor)

Resulted in [selection.csv](./selection.csv).

```sql
with uniq_msgs as (
    select msgs.cid                     as message_cid,
           actors.code                  as receiver_code,
           msgs.method                  as method_num,
           receipts.exit_code           as exit_code,
           b.height                     as height,
           b.cid                        as block_cid,
           row_number()
           over (partition by msgs.cid) as uniq_rn
    from public.messages as msgs -- join messages, with their blocks, their actor types, and receipts.
             join public.block_messages as block_msgs on msgs.cid = block_msgs.message
             join public.block_headers as b on b.cid = block_msgs.block
             join public.actors as actors
                  on msgs.to = actors.id and actors.state_root = b.parent_state_root -- this is not precise, but actor types are immutable, so it'll suffice
             join public.receipts as receipts on msgs.cid = receipts.message
    where b.height >= 65000
     and b.height <= 69000 -- between 65000 and 69000, inclusive.
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


## SQL query (against chainwatch)

Above height 50000. Chainwatch was synced only up to 60063. Not used to
generate this batch.

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
    where blocks.height >= 50000 -- process only heights above epoch 50000; chainwatch only synced up to 60063.
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