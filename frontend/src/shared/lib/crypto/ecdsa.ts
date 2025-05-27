import { ec as EC } from 'elliptic';
import CryptoJS from 'crypto-js';

const ec = new EC('secp256k1');

export interface ECDSAKeyPair {
  privateKey: string;
  publicKey: string;
}

export class ECDSAService {
  private keyPair: any;

  constructor() {
    this.generateKeyPair();
  }

  generateKeyPair(): ECDSAKeyPair {
    this.keyPair = ec.genKeyPair();
    const privateKey = this.keyPair.getPrivate('hex');
    const publicKey = this.keyPair.getPublic('hex');
    
    return {
      privateKey,
      publicKey,
    };
  }

  getPublicKey(): string {
    return this.keyPair.getPublic('hex');
  }

  getPrivateKey(): string {
    return this.keyPair.getPrivate('hex');
  }

  sign(data: string): string {
    const hash = CryptoJS.SHA256(data).toString();
    const signature = this.keyPair.sign(hash);
    return signature.toDER('hex');
  }

  verify(data: string, signature: string, publicKeyHex: string): boolean {
    try {
      const hash = CryptoJS.SHA256(data).toString();
      const publicKey = ec.keyFromPublic(publicKeyHex, 'hex');
      return publicKey.verify(hash, signature);
    } catch (error) {
      return false;
    }
  }

  static fromPrivateKey(privateKey: string): ECDSAService {
    const service = new ECDSAService();
    service.keyPair = ec.keyFromPrivate(privateKey, 'hex');
    return service;
  }

  static fromPublicKey(publicKey: string): ECDSAService {
    const service = new ECDSAService();
    service.keyPair = ec.keyFromPublic(publicKey, 'hex');
    return service;
  }

  /**
   * Generate a new ECDSA key pair (static method)
   */
  static generateStaticKeyPair(): ECDSAKeyPair {
    const keyPair = ec.genKeyPair();
    
    return {
      privateKey: keyPair.getPrivate('hex'),
      publicKey: keyPair.getPublic('hex')
    };
  }

  /**
   * Sign a message with ECDSA private key (static method)
   */
  static signStatic(message: string, privateKey: string): string {
    try {
      const keyPair = ec.keyFromPrivate(privateKey, 'hex');
      const hash = CryptoJS.SHA256(message).toString();
      const signature = keyPair.sign(hash);
      
      return signature.toDER('hex');
    } catch (error) {
      console.error('ECDSA signing error:', error);
      throw new Error('Failed to sign message with ECDSA');
    }
  }

  /**
   * Verify a signature with ECDSA public key (static method)
   */
  static verifyStatic(message: string, signature: string, publicKey: string): boolean {
    try {
      const keyPair = ec.keyFromPublic(publicKey, 'hex');
      const hash = CryptoJS.SHA256(message).toString();
      
      return keyPair.verify(hash, signature);
    } catch (error) {
      console.error('ECDSA verification error:', error);
      return false;
    }
  }

  /**
   * Validate a public key
   */
  static isValidPublicKey(publicKey: string): boolean {
    try {
      ec.keyFromPublic(publicKey, 'hex');
      return true;
    } catch {
      return false;
    }
  }

  /**
   * Validate a private key
   */
  static isValidPrivateKey(privateKey: string): boolean {
    try {
      ec.keyFromPrivate(privateKey, 'hex');
      return true;
    } catch {
      return false;
    }
  }
}
