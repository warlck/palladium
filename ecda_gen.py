# might require pip install ecdsa
from ecdsa import SigningKey, SECP256k1
sk = SigningKey.generate(curve=SECP256k1) # uses SECP256k1
vk = sk.get_verifying_key()

print('This is the secret key {} \nThis is the public key {}'.format(sk.to_string().hex(), vk.to_string().hex()))