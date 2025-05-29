# Text-based Clash Royale (TCR) - Application PDU Description

This document describes the Protocol Data Units (PDUs) used for communication between TCR clients and the server.

## Message Framing

TCR uses length-prefixed framing for all messages:

1. Each message is first serialized to JSON
2. A 4-byte header containing the length of the JSON message is prepended
3. The receiver reads the 4-byte length, then reads that many bytes to get the complete message

## Message Structure

All messages follow this general structure:

```json
{
  "type": "MESSAGE_TYPE",
  "payload": {
    // Message-specific payload fields
  }
}
```

## Message Types

### Connection Management

#### LOGIN_REQUEST
Sent by client to log in to the server.

```json
{
  "type": "LOGIN_REQUEST",
  "payload": {
    "username": "PlayerName",
    "password": "PlayerPassword"
  }
}
```

#### REGISTER_REQUEST
Sent by client to register a new account on the server.

```json
{
  "type": "REGISTER_REQUEST",
  "payload": {
    "username": "NewPlayerName",
    "password": "NewPlayerPassword"
  }
}
```

#### LOGIN_RESPONSE
Sent by server in response to a login request.

```json
{
  "type": "LOGIN_RESPONSE",
  "payload": {
    "success": true, // or false
    "message": "Successfully logged in as PlayerName",
    "playerId": "PlayerName" // Included if success is true
  }
}
```

#### REGISTER_RESPONSE
Sent by server in response to a registration request.

```json
{
  "type": "REGISTER_RESPONSE",
  "payload": {
    "success": true, // or false
    "message": "Successfully registered PlayerName"
  }
}
```

#### ERROR_NOTIFICATION
Sent by server to notify client of an error (e.g., login failure, invalid command, etc.).

```json
{
  "type": "ERROR_NOTIFICATION",
  "payload": {
    "errorMessage": "Description of the error (e.g., Invalid password)"
  }
}
```

### Game Management

#### DEPLOY_TROOP_COMMAND
Sent by client to deploy a troop.

```json
{
  "type": "DEPLOY_TROOP_COMMAND",
  "payload": {
    "troopName": "Knight",
    "targetTowerID": "OpponentUsername_GUARD1"
  }
}
```

#### SKIP_TURN_COMMAND
Sent by client to skip their current turn.

```json
{
  "type": "SKIP_TURN_COMMAND",
  "payload": {} // No payload needed
}
```

#### GAME_START_NOTIFICATION
Sent by server to notify clients that a game is starting.
Payload includes opponent's username and the initial state for the receiving player.

```json
{
  "type": "GAME_START_NOTIFICATION",
  "payload": {
    "opponentUsername": "OpponentPlayer",
    "yourPlayerInfo": { /* PlayerState object for the recipient */ },
    "gameMode": "SIMPLE"
  }
}
```

#### GAME_STATE_UPDATE
Sent by server to update clients on the current game state.
Includes the full state of both players, the current turn, and a log of the last action.

```json
{
  "type": "GAME_STATE_UPDATE",
  "payload": {
    "playerA": { /* PlayerState object for player A */ },
    "playerB": { /* PlayerState object for player B */ },
    "currentTurn": "PlayerName",
    "lastActionLog": "PlayerName deployed Knight..."
  }
}
```

#### ACTION_RESULT
Sent by server to notify the acting client of the result of their action (e.g., troop deployment, skip).

```json
{
  "type": "ACTION_RESULT",
  "payload": {
    "success": true, // or false
    "action": "Deploy Knight",
    "message": "Knight dealt 100 damage..."
  }
}
```

#### TURN_NOTIFICATION
Sent by server to notify the client whose turn it is now.

```json
{
  "type": "TURN_NOTIFICATION",
  "payload": {
    "currentTurnUsername": "PlayerName"
  }
}
```

#### GAME_OVER_NOTIFICATION
Sent by server to notify clients that the game is over.

```json
{
  "type": "GAME_OVER_NOTIFICATION",
  "payload": {
    "winnerUsername": "PlayerName", // Can be empty or "DRAW"
    "reason": "King Tower destroyed" // or "Player disconnected"
  }
}
```

<!-- Note: PlayerState and TowerState object structures are detailed in models.go -->
<!-- It's implied they are nested within payloads like GAME_START_NOTIFICATION and GAME_STATE_UPDATE -->