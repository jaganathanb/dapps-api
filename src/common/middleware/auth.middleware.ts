import passport from 'passport';
import httpStatus from 'http-status';
import { NextFunction, Request } from 'express';

import { roleRights } from '@config/roles';
import { User } from '@db/entity/User';

import { ApiError } from '../utils/api-error';

const verifyCallback =
  (
    req: Request,
    resolve: CallableFunction,
    reject: CallableFunction,
    requiredRights: string[],
  ) =>
  (err: ApiError, user: User) => {
    if (err || !user) {
      reject(new ApiError(httpStatus.UNAUTHORIZED, 'Please authenticate'));
    }
    req.user = user;

    if (requiredRights.length) {
      const userRights = roleRights.get(user.role);
      const hasRequiredRights = requiredRights.every(
        (requiredRight) => userRights?.includes(requiredRight),
      );
      if (!hasRequiredRights && req.params.userId !== user.id) {
        reject(new ApiError(httpStatus.FORBIDDEN, 'Forbidden'));
      }
    }

    resolve();
  };

const auth =
  (...requiredRights: string[]) =>
  async (req: Request, res: Response, next: NextFunction): Promise<void> =>
    new Promise((resolve, reject) => {
      // eslint-disable-next-line @typescript-eslint/no-unsafe-call
      passport.authenticate(
        'jwt',
        { session: false },
        verifyCallback(req, resolve, reject, requiredRights),
      )(req, res, next);
    })
      .then(() => next())
      .catch((err) => next(err));

export { auth };
