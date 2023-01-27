# Laboratoire 4 de SDR - Application de l'algorithme de Chang et Roberts

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

![graph](./docs/graph.png)

## Procédure de tests manuels

### Tests basiques


### Test 1
Faire un wave puis vérifier la réponse.

### Test 2
Faire un wave puis vérifier si la réponse est similaire sur 2 serveurs différents.

### Test 3
Faire deux wave de suite.

### Test 4
Faire un probe puis vérifier la réponse

### Test 5

### Test 6


## Implémentation

### Le client

Le client effectue une nouvelle connexion UDP à un serveur à chaque commande envoyée. 

Le client parse l'input en ligne de commande et crée un objet `Command` si l'input est valide. Il transforme ensuite cet objet en string JSON et l'envoie au serveur.

Quitter un client avec CTRL+C ou en envoyant la commande `quit` ferme la connexion UDP en cours et arrête le client gracieusement.

### Le serveur


### Points à améliorer
