# Batch 0004: Coverage boost

> first extracted around Wed Sep 30 15:00:00 UTC 2020 (tvx version: eb6191d0ffd01a7cf7f8544a31acf307b1799fb2)
>
> reextracted around Wed Oct 14 18:00:00 UTC 2020 (tvx version: https://github.com/filecoin-project/lotus/pull/4393)

This is a selection of 344 messages extracted from heights the Space Race chain
(which later transitioned to Ignition), up until height ~57000, with similar
methods to batch 0001, but with a larger timespan, thus picking up more variance
in samples. There may be significant coverage overlap with batch 0001.

## Message selection

Ten most recent messages a unique combination of
`(actor_code,method_num,exit_code)`, from the Space Race chain up until
height ~57000. 

## Unsuccessful extraction

The following messages were not extracted successfully due to mismatches in
local execution receipt vs. chain receipt. Once issues are fixed, a future batch
may incorporate these vectors. They appear in [selection.csv](./selection.csv),
but do not have corresponding vectors.

```
* message bafy2bzaceahrgvgtiz3zenvatp3zbjpyme3y7q6jqhora2n54nsxgzpacswra: message not found; precursors found: 285
* message bafy2bzacecnpanewojacbkns2ek5e7hpzs32z6ngiaq25cvo6dg6kgfsa4qyq: vector generation aborted (receipt mismatch)
* message bafy2bzacecziet2dfsollbjwqqa4yvrahhxjtjhknqnnjb6ynhkn5l2hbav7k: vector generation aborted (receipt mismatch)
* message bafy2bzaceb5brd6qy6wtlmjzbrc24lbi2ro5i22yvgl2rv4xxwzcbxz2kk2lw: vector generation aborted (receipt mismatch)
* message bafy2bzacecyctgiopr4r2ctirq7iipdycc4yzv4oljddkcgbdtn6vhnillef6: vector generation aborted (receipt mismatch)
* message bafy2bzaceam7a4wtwrwhrouvbewmuxyexlf2jf2jmbg33sy2aeigge6heulyk: vector generation aborted (receipt mismatch)
* message bafy2bzaceb4l6ugopvkjrwtz2rioxn5ba6fk25p42dvwsx772vwbq53loj7yo: vector generation aborted (receipt mismatch)
* message bafy2bzacediqx3kywu73ilquo2wpeimckv355unlk5iq4o4xujnryozrlrwna: vector generation aborted (receipt mismatch)
* message bafy2bzacec6qyurlpq7sygfkhvwvhx6e73ljz23yd3y5777uwftxm5hgv4mju: vector generation aborted (receipt mismatch)
* message bafy2bzacedptbjotdvonom7oh3kilknjwcxzx5s4jpqyeycyyuygwz4ja3s2q: vector generation aborted (receipt mismatch)
* message bafy2bzacechu6pabplwqzh7dqpz5w5pcpv6d4gj2jwxy6qsyib53le4ccw6kg: vector generation aborted (receipt mismatch)
* message bafy2bzacedsj4omiq25udyrlphhloxhzd6ptmvp7f5jer37xb5csag7c4vp3a: vector generation aborted (receipt mismatch)
* message bafy2bzaceb35avh6raml3r4o3rfecfsezoziyni2hbdp7dtotphpklxi7zsdo: message not found; precursors found: 528
* message bafy2bzacedyzaknevzsqzpod7b6aq5narof4av6wu3eei7xbe3g2trs2kajcu: vector generation aborted (receipt mismatch)
* message bafy2bzacearqalrvu3ukeiidtqh3gzbzokdqv7v27gv4ylzhxmkf5mfkycfn4: vector generation aborted (receipt mismatch)
* message bafy2bzacebyzu6s6mlbsge2iw23i422etzpsmncjh72otprwjj4jfeerpmuis: message not found; precursors found: 525
* message bafy2bzaced47iodrh6xktfh6toms254kqqctl4k46ejyiijaq4cry7rrl5mea: message not found; precursors found: 454
* message bafy2bzaceclhd3umoby6pipoefc3iqu2h26dtmfr2gzqilxuhugsw5kgrmvws: vector generation aborted (receipt mismatch)
* message bafy2bzacearlhgbt6srogzkyjkalvr25xtwh4qulykjixy3nkpscrun2gv65g: vector generation aborted (receipt mismatch)
* message bafy2bzacebf7norxehntwshdywagvit32rtpgqoqads6wasyevg5w2cz442eg: vector generation aborted (receipt mismatch)
* message bafy2bzaceb2jm3zw7dv3hupdm2fkvpouqkzs64wwv64xuc34x2urjlyf7ukng: vector generation aborted (receipt mismatch)
* message bafy2bzaceclbapzvrc6vpemtjqqmaakphqfoyhzt2dl6ocfu2mh6ox2sm4aii: vector generation aborted (receipt mismatch)
* message bafy2bzacecku6m7uesqkqtnhp4bww4be46wlerbskizslvp3ih335xdd2dwsg: vector generation aborted (receipt mismatch)
* message bafy2bzacebb2qpol6ukigma7abi5gg2w5sgargicpqzyjunhcaehruha76qcc: vector generation aborted (receipt mismatch)
* message bafy2bzaceaqbhmzoqy7reqnbhnit37jopg3a2yx5z4dtpzevqyery36sgxuao: vector generation aborted (receipt mismatch)
* message bafy2bzacebqzzjluln7mmngkocjqkvkvcjsnudm52lngwshyxzyjvz2he7x7e: vector generation aborted (receipt mismatch)
* message bafy2bzacea6ohve6z3eqcsyplzozdxif6xekrnw623adlvhjbessfq6mtmovc: vector generation aborted (receipt mismatch)
* message bafy2bzacecck6xz5xbhug3ezqj6wh3rmnhrnyhfg46pxlzbmyr5udm3wiofiq: vector generation aborted (receipt mismatch)
* message bafy2bzacedigegsxhagbsgy4hoxbpfoxvtf4n62qow7uswartb7nm3uqdag34: vector generation aborted (receipt mismatch)
* message bafy2bzacea5n4cssv63l4ej54j63t4tuv3zcpnjxkx5qzhif6l2bbqfp37ja6: vector generation aborted (receipt mismatch)
* message bafy2bzacea6tkskkncok5655e2pajmj7ltjbhny4vndf7bxlv26gs2h523jom: vector generation aborted (receipt mismatch)
```

## SQL query (against Chainwatch DB)

Up until height ~57000 (Chainwatch instance stopped processing after then).

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
