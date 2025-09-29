# Blockchain Directory

This directory is used by the Hardhat node service in Docker Compose to initialize a local Ethereum development network.

The Hardhat node will automatically:
1. Initialize a new Hardhat project if needed
2. Install required dependencies
3. Start a local Ethereum node on port 8545

## Smart Contracts

Smart contracts will be developed and deployed here in future phases of the project.

## Configuration

The Docker Compose service will create:
- `package.json` - Node.js project configuration
- `hardhat.config.js` - Hardhat configuration
- `contracts/` - Smart contract source files (Solidity)
- `scripts/` - Deployment and interaction scripts
- `test/` - Smart contract tests

## Network Details

- Chain ID: 31337 (Hardhat default)
- RPC URL: http://localhost:8545
- WebSocket URL: ws://localhost:8545
- Pre-funded accounts with test ETH available