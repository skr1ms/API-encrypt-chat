import CryptoJS from 'crypto-js';
import forge from 'node-forge';

export interface RSAKeyPair {
  privateKey: string;
  publicKey: string;
}

export class RSAService {
  private keyPair: forge.pki.rsa.KeyPair | null = null;

  generateKeyPair(keySize: number = 2048): RSAKeyPair {
    this.keyPair = forge.pki.rsa.generateKeyPair(keySize);
    
    const privateKeyPem = forge.pki.privateKeyToPem(this.keyPair.privateKey);
    const publicKeyPem = forge.pki.publicKeyToPem(this.keyPair.publicKey);
    
    return {
      privateKey: privateKeyPem,
      publicKey: publicKeyPem,
    };
  }

  encrypt(data: string, publicKeyPem: string): string {
    const publicKey = forge.pki.publicKeyFromPem(publicKeyPem);
    const encrypted = publicKey.encrypt(data, 'RSA-OAEP');
    return forge.util.encode64(encrypted);
  }

  decrypt(encryptedData: string, privateKeyPem: string): string {
    const privateKey = forge.pki.privateKeyFromPem(privateKeyPem);
    const encrypted = forge.util.decode64(encryptedData);
    const decrypted = privateKey.decrypt(encrypted, 'RSA-OAEP');
    return decrypted;
  }

  sign(data: string, privateKeyPem: string): string {
    const privateKey = forge.pki.privateKeyFromPem(privateKeyPem);
    const md = forge.md.sha256.create();
    md.update(data, 'utf8');
    const signature = privateKey.sign(md);
    return forge.util.encode64(signature);
  }

  verify(data: string, signature: string, publicKeyPem: string): boolean {
    try {
      const publicKey = forge.pki.publicKeyFromPem(publicKeyPem);
      const md = forge.md.sha256.create();
      md.update(data, 'utf8');
      const signatureBytes = forge.util.decode64(signature);
      return publicKey.verify(md.digest().bytes(), signatureBytes);
    } catch (error) {
      return false;
    }
  }

  static generateKeyPair(keySize: number = 2048): RSAKeyPair {
    const keyPair = forge.pki.rsa.generateKeyPair(keySize);
    
    const privateKeyPem = forge.pki.privateKeyToPem(keyPair.privateKey);
    const publicKeyPem = forge.pki.publicKeyToPem(keyPair.publicKey);
    
    return {
      privateKey: privateKeyPem,
      publicKey: publicKeyPem,
    };
  }

  static fromPrivateKey(privateKeyPem: string): RSAService {
    const service = new RSAService();
    const privateKey = forge.pki.privateKeyFromPem(privateKeyPem);
    const publicKey = forge.pki.setRsaPublicKey(privateKey.n, privateKey.e);
    service.keyPair = { privateKey, publicKey };
    return service;
  }
}
