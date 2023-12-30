import express, { Response } from 'express';
import debug from 'debug';

import { Jwt } from '@common/types/jwt';

import { PermissionLevel } from './common.permissionlevel.enum';

const log: debug.IDebugger = debug('app:common-permission-middleware');

const minimumPermissionLevelRequired =
  (requiredPermissionLevel: PermissionLevel) =>
  (
    req: express.Request,
    res: express.Response,
    next: express.NextFunction,
  ): void => {
    try {
      const userPermissionLevel = parseInt(
        (res.locals.jwt as Jwt).permissionLevel,
      );
      if (userPermissionLevel > Number(requiredPermissionLevel)) {
        next();
      } else {
        res.status(403).send();
      }
    } catch (e) {
      log(e);
    }
  };

const onlySameUserOrAdminCanDoThisAction = (
  req: express.Request,
  res: express.Response,
  next: express.NextFunction,
): void | Response => {
  const userPermissionLevel = parseInt((res.locals.jwt as Jwt).permissionLevel);
  if (
    req.params &&
    req.params.userId &&
    req.params.userId === (res.locals.jwt as Jwt).userId
  ) {
    return next();
  } else {
    if (userPermissionLevel > Number(PermissionLevel.AdminPermission)) {
      return next();
    } else {
      return res.status(403).send();
    }
  }
};

const onlyAdminCanDoThisAction = (
  req: express.Request,
  res: express.Response,
  next: express.NextFunction,
): void | Response => {
  const userPermissionLevel = parseInt((res.locals.jwt as Jwt).permissionLevel);
  if (userPermissionLevel > Number(PermissionLevel.AdminPermission)) {
    return next();
  } else {
    return res.status(403).send();
  }
};

export {
  onlyAdminCanDoThisAction,
  minimumPermissionLevelRequired,
  onlySameUserOrAdminCanDoThisAction,
};
