# Laboratoire 4 de SDR - Comptage du nombre d'occurrences de lettres dans un texte dans un système distribué

## Auteurs

Lazar Pavicevic et Jonathan Friedli

## Contexte

Ce projet est réalisé dans le cadre du cours de Systèmes Distribués et Répartis (SDR) de la HEIG-VD.

Dans ce laboratoire, nous implémentons deux algorithmes: l'algorithme ondulatoire et l'algorithme sondes et échos afin de compter de façon distribuée le nombre d’occurrences de lettres dans un texte grâce à un réseau de serveurs. Le réseau est agencé comme un graphe. Toutes les connexions sont réalisées en UDP.

## Utilisation du programme

L'application contient deux exécutables : un pour le serveur et un pour le client.
### Pour lancer un serveur:

Le serveur a besoin d'un entier en argument qui représente la clé des maps présentes dans son fichier de configuration. Ces maps indiquent l'adresse de tout les autres serveurs composant le réseau.

```bash
# A la racine du projet

# Lancement du serveur n°1 en mode race
go run -race cmd/server/main.go 1
```

### Pour lancer un client:

Le client n'a pas besoins d'arguments pour être lancé.

```bash
# A la racine du projet

# Lancement d'un client en mode race
go run -race cmd/client/main.go
```

### Commandes disponibles:

```bash

# Commande demandant le traitement d'un texte avec l'algorithme ondulatoire
wave <word>

# Commande demandant le traitement d'un texte avec l'algorithme sondes et échos en spécifiant le serveur racine
probe <word> <server number>

# Commande demandant le résultat du dernier traitement effectué
# Ondulatoire: Tout les serveurs peuvent répondre
# Sondes et échos: Seul le serveur racine peut répondre
ask <server number>

# Commande permettant de quitter le client
quit
```

# Les tests

Il n'était pas demandé d'effectuer des tests unitaires et automatisés pour ce laboratoire. Il n'était pas non plus demandé de simuler des ralentissements ou des paquets perdus. Nous avons donc effectué des tests manuels pour vérifier le bon fonctionnement de notre application sans utiliser de mode debug simulant ces dégradations.

Tous nos tests sont effectués avec la configuration fournie dans les fichiers de `config.json` de chaque exécutable. Dans le fichier `config.json` du serveur, nous avons ajouté une liste d'adjacence pour identifier les serveurs voisins de chaque serveur.

Le graphe logique de notre réseau est le suivant:

![graph](./docs/graph.png)

Nous avons ajouté un cycle simple entre P0-P1-P2 pour tester la détection de cycle. En rouge, nous retrouvons l'unique lettre qui sera traitée par le serveur.

## Procédure de tests manuels


### Test n°1
On commence par faire la commande `ask` sur le serveur 0

Input du client:
```bash
ask 0
```

Résultat attendu:
Le server 0 nous répond qu'il n'y a pas encore eu de texte traité.

### Test n°2
On fait la commande `wave` avec le mot "pomme" puis on fait la commande `ask` sur le serveur 0.

Input du client:
```bash
wave pomme
ask 0
```

Résultat attendu:
Le server 0 nous montre le nombre d'occurrences de chaque lettre dans le mot "pomme".

### Test n°3
On fait la commande `probe` avec le mot "pomme" sur le serveur 0 puis on fait la commande `ask` sur le serveur 1.

Input du client:
```bash
probe pomme 0
```

Résultat attendu:
Le server 0 nous montre le nombre d'occurrences de chaque lettre dans le mot "pomme".

### Test n°4
Faire un wave avec un mot comprenant des lettres non traitées.
Input du client:
```bash
wave vache
ask 0
```

Résultat attendu:
La seule lettre traitée étant la lettre "e", le server 0 nous montre le nombre d'occurrences de cette lettre dans le mot "vache".

### Test n°5
Faire un wave avec un mot comprenant des lettres non traitées.
Input du client:
```bash
probe banc 4
```

Résultat attendu:
Aucune lettre n'est traitée, le server 4 nous répond qu'aucune lettre traitée n'a été trouvée.

### Test n°6
Faire un wave puis vérifier si la réponse est similaire sur 2 serveurs différents.

Input du client:
```bash
wave tombe
ask 0
ask 4
```

Résultat attendu:
On remarque que l'ordre des lettres n'est pas pareil mais le compte de chaque lettre est équivalent.

### Test n°7
Faire un probe puis vérifier si la réponse est similaire sur 2 serveurs différents.

Input du client:
```bash
probe pomte 0
ask 0
ask 4
```

Résultat attendu:
Le serveur P0 nous donne le bon compte de lettre mais le serveur P4 nous dit qu'aucun texte n'a été traité car ce dernier est un processus enfant dans cette itération.



## Implémentation

### Le client

Le client effectue une nouvelle connexion UDP à un serveur à chaque commande envoyée. 

Le client parse l'input en ligne de commande et crée un objet `Command` si l'input est valide. Il transforme ensuite cet objet en string JSON et l'envoie au serveur.

Quitter un client avec CTRL+C ou en envoyant la commande `quit` ferme la connexion UDP en cours et arrête le client gracieusement.

### Le serveur


### Points à améliorer
