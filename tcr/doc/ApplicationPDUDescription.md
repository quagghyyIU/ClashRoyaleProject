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
    "username": "PlayerName"
  }
}
```

#### LOGIN_RESPONSE
Sent by server in response to a login request.

```json
{
  "type": "LOGIN_RESPONSE",
  "payload": {
    "success": true/false,
    "message": "Success or error message",
    "playerId": "PlayerID" // Only included if success is true
  }
}
```

#### ERROR_NOTIFICATION
Sent by server to notify client of an error.

```json
{
  "type": "ERROR_NOTIFICATION",
  "payload": {
    "errorMessage": "Description of the error"
  }
}
```

### Game Management (Phase 3)

These messages will be implemented in Phase 3 of the project:

#### DEPLOY_TROOP_COMMAND
Sent by client to deploy a troop.

```json
{
  "type": "DEPLOY_TROOP_COMMAND",
  "payload": {
    "troopName": "TroopName",
    "targetTowerID": "OpponentUsername_TOWERTYPE"
  }
}
```

#### GAME_START_NOTIFICATION
Sent by server to notify clients that a game is starting.

```json
{
  "type": "GAME_START_NOTIFICATION",
  "payload": {
    "opponentUsername": "OpponentName",
    "yourPlayerInfo": {
      // Player info
    },
    "initialGameState": {
      // Game state
    },
    "gameMode": "SIMPLE" // or "ENHANCED"
  }
}
```

#### GAME_STATE_UPDATE
Sent by server to update clients on the current game state.

```json
{
  "type": "GAME_STATE_UPDATE",
  "payload": {
    "playerA": {
      // Player A state
    },
    "playerB": {
      // Player B state
    },
    "currentTurn": "PlayerName",
    "lastActionLog": "Action description" // Optional
  }
}
```

#### ACTION_RESULT
Sent by server to notify client of the result of their action.

```json
{
  "type": "ACTION_RESULT",
  "payload": {
    "success": true/false,
    "action": "Action description",
    "message": "Result description"
  }
}
```

#### TURN_NOTIFICATION
Sent by server to notify client that it's their turn.

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
    "winnerUsername": "PlayerName",
    "reason": "Reason for game end"
  }
}
```