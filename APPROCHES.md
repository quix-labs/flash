# Approche 1 [TRIGGER DISTINCTS]

## Bootstraping

- Generation d'un nom unique:
    - Reference unique+action (insert,update,delete,truncate)

- Création
    - Création d'une trigger function -> pg_notify(nom_unique)
    - Créer un trigger FOR EACH ROW qui appelle la ttrigger function precedante

- Suppression:
    - Suppression de la trigger function dédié à l'action (a partir du nom unique comme reference) CASCADE
    - Comme cascader, le trigger se supprimera également

## Reception d'event

- Chaque event est reconnaissable par un nom unique qu'il emet dans pg_notify.
    - Comme nous n'avons de trigge que pour les données demandé, on retourne l'event vers le callback dans tous les cas

# Approche2 [TRIGGER GLOBAL UPDATE/DELETE/INSERT + TRIGGER TRUNCATE]

## Bootstraping

- Generation d'un nom unique:
    - Si TRUNCATE: Reference unique+truncate -> ex: flash_posts_truncate
    - Aussi non Reference unique+other -> ex: flash_posts_other

- Création
    - Si TRUNCATE -> CREATE TRIGGER ON ... BEFORE TRUNCATE FOR EACH STATEMENT ...
    - Aussi non:
        - Si trigger global déjà existant -> on ignore
        - Si aucun trigger global enregistré on le crée -> CREATE TRIGGER ON ... BEFORE UPDATE,DELETE,INSERT FOR EACH
          STATEMENT ...
            - Parcours old_table et new_table -> pour chaque entrée appelle pg_notify en passant TG_OP

## Reception d'event

Dans ce cas nous recevront des event non ecoutés

A nous de verifier si l'event recu fait parti de la liste d'event ecouté.

- Si oui on l'envoie vers le callback
- Si non on on l'ignore

# Approche3 [TRIGGER GLOBAL UPDATE/DELETE/INSERT + TRIGGER TRUNCATE] VERSION BATCH

## Bootstraping

- Comme l'approche 2 mais au lieu d'appeller pg_notify pour chaque ligne, on genere un tableau json et on envoie le
  payload complet une seuls fois

## Reception d'event

- Comme l'approche 2 mais si on recoit le payload, on le decode et on parcoure chaque entrée pour envoyer un evenement a
  chaque entrée

# Approche4 [WAL REPLICATION - A PEAUFINER]

## Bootstraping

- CREATION:
    - En interne on l'ajoute a notre liste d'evenement a suivre
- SUPPRESSION:
    - Si existant, en interne, on le supprime de notre liste d'evenement a suivre

## Reception d'event

Dans tous les cas on recevra tous les evenements, écoutés ou pas.

- On parse la log de replication, on extrait l'operation et la table + ...
    - Detection de l'event (INSERT,UPDATE,DELETE,TRUNCATE,...)
        - Si n'existe pas, en interne, dans notre liste d'evenement a suivre -> ignore
        - Aussi non -> emission vers le callback
        
# Approche 5 [CUSTOM PLUGIN - REFLEXION NECESSAIRE]

## Bootstraping

- CREATION:
    - Call custom fonction pour ecouter
- SUPPRESSION:
    - Call custom fonction pour ne plus ecouter

## Reception d'event

- On recupere l'evenement emis
    - On l'envoie vers le callback