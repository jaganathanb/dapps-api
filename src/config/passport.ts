import {
  Strategy as JwtStrategy,
  ExtractJwt,
  VerifiedCallback,
} from 'passport-jwt';

import { readById } from '../users/users.repository';

import { defaultConfig as config } from './base';
import { TokenTypes } from './tokens';

const jwtOptions = {
  secretOrKey: config.jwt.secret,
  jwtFromRequest: ExtractJwt.fromAuthHeaderAsBearerToken(),
};

const jwtVerify = async (
  payload: { type: string; sub: string },
  done: VerifiedCallback,
) => {
  try {
    if (payload.type !== TokenTypes.ACCESS) {
      throw new Error('Invalid token type');
    }
    const user = await readById(payload.sub);
    if (!user) {
      return done(null, false);
    }
    done(null, user);
  } catch (error) {
    done(error, false);
  }
};

const jwtStrategy = new JwtStrategy(jwtOptions, jwtVerify);

export { jwtStrategy };
