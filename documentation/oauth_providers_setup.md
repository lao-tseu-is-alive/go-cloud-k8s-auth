# Guide de Configuration des Fournisseurs OAuth2

Ce guide explique en détail comment enregistrer votre application auprès des différents fournisseurs d'identité (Google, GitHub, et Microsoft) afin d'obtenir les identifiants requis pour le fonctionnement du service `go-cloud-k8s-auth`.

---

## 💡 Principe de fonctionnement du flux OAuth2 dans ce projet

1. **Demande de connexion** : L'utilisateur clique sur un bouton dans le frontend (ex: GitHub).
2. **Génération de l'URL** : Le frontend appelle `POST /goapi/v1/auth/start` sur notre serveur Go. Le serveur construit l'URL de connexion du fournisseur avec un paramètre `redirect_uri` configuré dans votre fichier `.env` (ex: `http://localhost:9090/`) et un jeton unique de sécurité `state` (anti-CSRF).
3. **Consentement** : L'utilisateur est redirigé vers l'écran d'authentification du fournisseur (ex: GitHub Login).
4. **Redirection de retour** : Une fois connecté, le fournisseur renvoie le navigateur de l'utilisateur vers notre frontend à l'URL de redirection avec deux paramètres : `code` (le code d'autorisation temporaire) et `state` (le jeton anti-CSRF).
   * *Exemple d'URL de redirection :* `http://localhost:9090/?code=12345&state=abcde`
5. **Échange final (Callback)** : Notre application frontend (servie à la racine `/`) intercepte ces paramètres et fait un appel en tâche de fond en faisant un `POST /goapi/v1/auth/callback` au serveur. Le serveur valide le jeton `state`, échange le `code` temporaire contre un jeton d'accès réel auprès du fournisseur, récupère le profil de l'utilisateur, l'enregistre en base de données, génère et retourne un token JWT local.

---

## 🐱 1. Configuration pour GitHub

### Étapes d'enregistrement :
1. Connectez-vous à votre compte **GitHub**.
2. Allez dans vos **Settings** (Paramètres de profil).
3. Dans la barre latérale gauche, cliquez sur **Developer Settings** (tout en bas).
4. Cliquez sur **OAuth Apps**, puis sur **New OAuth App** (ou *Register a new application*).
5. Remplissez le formulaire :
   * **Application name** : `Go-Cloud-Auth (Dev)`
   * **Homepage URL** : `http://localhost:9090`
   * **Authorization callback URL** : `http://localhost:9090/` (⚠️ *Très important: inclure le slash final, et ne pointez pas vers le backend* `/goapi/v1/auth/callback` *car le navigateur ferait un GET dessus, ce qui produirait une erreur 404*).
6. Cliquez sur **Register application**.
7. Sur la page récapitulative, copiez le **Client ID**.
8. Cliquez sur **Generate a new client secret** et copiez immédiatement la valeur générée (elle ne s'affichera qu'une seule fois).

### Variables à ajouter dans le fichier `.env` :
```env
OAUTH_GITHUB_CLIENT_ID="votre_client_id_github"
OAUTH_GITHUB_CLIENT_SECRET="votre_client_secret_github"
OAUTH_GITHUB_REDIRECT_URL="http://localhost:9090/"
```

---

## 🌐 2. Configuration pour Google Cloud

### Étapes d'enregistrement :
1. Connectez-vous à la [Console Google Cloud](https://console.cloud.google.com/).
2. Créez un nouveau projet (ou sélectionnez-en un existant) via le sélecteur de projet en haut à gauche.
3. Dans la barre de recherche ou le menu de gauche, accédez à **API et services > Écran de consentement OAuth**.
4. Configurez l'écran de consentement :
   * Choisissez le type d'utilisateur **Externe** (External), puis cliquez sur **Créer**.
   * Renseignez les informations obligatoires de l'application (Nom de l'application, e-mail d'assistance, coordonnées du développeur).
   * À l'étape des **Champs d'application** (Scopes), ajoutez les scopes requis : `.../auth/userinfo.email`, `.../auth/userinfo.profile`, et `openid`.
   * Ajoutez vos adresses e-mail de test dans la section **Utilisateurs de test** (nécessaire tant que l'application n'est pas publiée).
5. Dans le menu de gauche, cliquez sur **Identifiants** (Credentials).
6. Cliquez sur **+ Créer des identifiants**, puis sélectionnez **ID client OAuth**.
7. Configurez l'ID client :
   * **Type d'application** : `Application Web`
   * **Nom** : `Go-Cloud-Auth (Dev)`
   * **Origines JavaScript autorisées** : `http://localhost:9090`
   * **URI de redirection autorisés** : `http://localhost:9090/` (⚠️ *Inclure le slash final*).
8. Cliquez sur **Créer**, puis copiez l'**ID client** et le **Code secret du client** dans la boîte de dialogue qui s'affiche.

### Variables à ajouter dans le fichier `.env` :
```env
OAUTH_GOOGLE_CLIENT_ID="votre_client_id_google"
OAUTH_GOOGLE_CLIENT_SECRET="votre_client_secret_google"
OAUTH_GOOGLE_REDIRECT_URL="http://localhost:9090/"
```

---

## 💻 3. Configuration pour Microsoft AD (Entra ID)

### Étapes d'enregistrement :
1. Connectez-vous au [Portail Azure](https://portal.azure.com/).
2. Dans la barre de recherche supérieure, tapez et sélectionnez **Microsoft Entra ID** (anciennement Azure Active Directory).
3. Dans le menu de gauche, cliquez sur **Inscriptions d'applications** (App registrations), puis sur **Nouvelle inscription** (New registration).
4. Renseignez le formulaire d'inscription :
   * **Nom** : `Go-Cloud-Auth (Dev)`
   * **Types de comptes pris en charge** : Sélectionnez *Comptes dans n'importe quel annuaire organisationnel (tout annuaire Microsoft Entra ID - multilocataire) et comptes Microsoft personnels (par exemple, Skype, Xbox)* afin de permettre la connexion à la fois pour les comptes pro/étudiants et grand public.
   * **URI de redirection** : Dans le menu déroulant, sélectionnez **Web**, puis saisissez `http://localhost:9090/`.
5. Cliquez sur **S'inscrire** (Register).
6. Dans l'onglet **Vue d'ensemble** (Overview), copiez la valeur de l'**ID d'application (client)**.
7. Dans le menu de gauche, allez dans **Certificats et secrets** (Certificates & secrets).
8. Cliquez sur **Nouveau secret client** (New client secret) :
   * Saisissez une description (ex: `dev-secret`).
   * Choisissez une date d'expiration (ex: 180 jours).
   * Cliquez sur **Ajouter**.
9. Copiez immédiatement la valeur présente dans la colonne **Valeur** (⚠️ *Attention à ne pas copier l'ID de secret, copiez bien la valeur elle-même*).

### Variables à ajouter dans le fichier `.env` :
```env
OAUTH_MICROSOFT_CLIENT_ID="votre_application_id_microsoft"
OAUTH_MICROSOFT_CLIENT_SECRET="votre_valeur_de_secret_microsoft"
OAUTH_MICROSOFT_REDIRECT_URL="http://localhost:9090/"
```

---

## 📋 Résumé pour le fichier `.env` local

Une fois toutes les étapes terminées, votre fichier `.env` de développement doit comporter les sections configurées comme suit :

```env
# URL d'écoute et de redirection (Frontend local)
PORT=9090

# Configuration PostgreSQL locale
DB_DRIVER=postgres
DB_HOST=127.0.0.1
DB_PORT=5432
DB_NAME=go_cloud_auth
DB_USER=votre_utilisateur_pg
DB_PASSWORD=votre_mot_de_pass_pg
DB_SSL_MODE=disable

# Configuration JWT locale
JWT_SECRET="UneCleSuperSecuriseeEtTresLongueDeVotreChoix"

# GitHub OAuth
OAUTH_GITHUB_CLIENT_ID="votre_client_id_github"
OAUTH_GITHUB_CLIENT_SECRET="votre_client_secret_github"
OAUTH_GITHUB_REDIRECT_URL="http://localhost:9090/"

# Google OAuth
OAUTH_GOOGLE_CLIENT_ID="votre_client_id_google"
OAUTH_GOOGLE_CLIENT_SECRET="votre_client_secret_google"
OAUTH_GOOGLE_REDIRECT_URL="http://localhost:9090/"

# Microsoft OAuth
OAUTH_MICROSOFT_CLIENT_ID="votre_application_id_microsoft"
OAUTH_MICROSOFT_CLIENT_SECRET="votre_valeur_de_secret_microsoft"
OAUTH_MICROSOFT_REDIRECT_URL="http://localhost:9090/"
```
