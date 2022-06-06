# Protocole de distribution d’ordonnancement pour réseaux de capteurs sans fil 

Les dossiers contenus dans ce répertoire contienne les implémentations du protocole développé dans le cadre de mon mémoire sur la distribution d'ordonnancement dans des réseaux de capteurs sans fil.

Les dossiers contiennent les implémentations suivantes.

## Côté client
Le code des clients est contenus dans les dossiers :
 - `common` : contient le code commun aux clients et au nœud de bordure.
 - `client` : contient le code des nœuds n'étant pas le nœud de bordure.
 - `proxy` : contient le code du nœud de bordure (aussi appelé proxy).
 - `simulations` : contient le script Javascript permettant de récolter des informations des nœuds via Cooja.

Ce code fonctionne avec le système d'exploitation Contiki-NG. L'implémentation de Contiki-NG n'est pas incluse dans le code fourni mais peut être téléchargé [ici](https://github.com/contiki-ng/contiki-ng).
La version de Contiki-NG spécifiquement utilisé pour ce mémoire ainsi que le code de l'implémentation peut aussi être téléchargé depuis [GitHub](https://github.com/FlorentCollin/contiki-ng).

## Côté server
Le code est contenu dans le dossier `server`. L'implémentation utilise uniquement la librairie standard de Go et aucune dépendance ne doit donc être téléchargé. Le code a été testé en utilisant Go 1.18.2.

Le dossier `utils` contient des utilitaires qui ont servi pour faire tourner les différentes simulations montrées dans ce mémoire. Notamment, le script bash `run-simulation` permet de lancer les simulations via l'utilisation de Docker pour Cooja et Kitty pour le multiplexing de terminal.

## Statistiques des simulations
Finalement, le dossier `server-simulations` contient les statistiques des simulations utilisées pour l'évaluation du protocole.
