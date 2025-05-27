import CryptoJS from 'crypto-js';
import { ec as EC } from 'elliptic';

const ec = new EC('secp256k1');

export interface KeyPair {
  privateKey: string;
  publicKey: string;
}

export class ECDHService {
  private keyPair: any;

  constructor() {
    this.generateKeyPair();
  }

  generateKeyPair(): KeyPair {
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

  computeSharedSecret(otherPublicKey: string): string {
    const otherKey = ec.keyFromPublic(otherPublicKey, 'hex');
    const sharedSecret = this.keyPair.derive(otherKey.getPublic());
    return sharedSecret.toString(16);
  }

  encryptMessage(message: string, sharedSecret: string): string {
    const key = CryptoJS.SHA256(sharedSecret).toString();
    const encrypted = CryptoJS.AES.encrypt(message, key).toString();
    return encrypted;
  }

  decryptMessage(encryptedMessage: string, sharedSecret: string): string {
    const key = CryptoJS.SHA256(sharedSecret).toString();
    const decrypted = CryptoJS.AES.decrypt(encryptedMessage, key);
    return decrypted.toString(CryptoJS.enc.Utf8);
  }

  static fromPrivateKey(privateKey: string): ECDHService {
    const service = new ECDHService();
    service.keyPair = ec.keyFromPrivate(privateKey, 'hex');
    return service;
  }

  /**
   * Generate a new ECDH key pair (static method)
   */
  static generateKeyPair(): KeyPair {
    const keyPair = ec.genKeyPair();
    
    return {
      privateKey: keyPair.getPrivate('hex'),
      publicKey: keyPair.getPublic('hex')
    };
  }

  /**
   * Compute shared secret (static method)
   */
  static computeSharedSecret(privateKey: string, otherPublicKey: string): string {
    const keyPair = ec.keyFromPrivate(privateKey, 'hex');
    const otherKey = ec.keyFromPublic(otherPublicKey, 'hex');
    const sharedSecret = keyPair.derive(otherKey.getPublic());
    return sharedSecret.toString(16);
  }

  /**
   * Encrypt message using shared secret (static method)
   */
  static encryptMessage(message: string, sharedSecret: string): string {
    const key = CryptoJS.SHA256(sharedSecret).toString();
    const encrypted = CryptoJS.AES.encrypt(message, key).toString();
    return encrypted;
  }

  /**
   * Decrypt message using shared secret (static method)
   */
  static decryptMessage(encryptedMessage: string, sharedSecret: string): string {
    const key = CryptoJS.SHA256(sharedSecret).toString();
    const decrypted = CryptoJS.AES.decrypt(encryptedMessage, key);
    return decrypted.toString(CryptoJS.enc.Utf8);
  }
}
