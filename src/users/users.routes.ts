import express from 'express';
import { body } from 'express-validator';

import {
  PermissionLevel,
  minimumPermissionLevelRequired,
  onlySameUserOrAdminCanDoThisAction,
  verifyBodyFieldsErrors,
} from '@common/middleware';

import {
  createUser,
  getUserById,
  listUsers,
  patch,
  put,
  removeUser,
} from './users.controller';
import {
  validatePatchEmail,
  validateRequiredUserBodyFields,
  validateSameEmailBelongToSameUser,
  validateSameEmailDoesntExist,
  validateUserExists,
} from './users.middleware';

const router = express.Router();

router
  .route(`/users`)
  .get(listUsers)
  .post(
    validateRequiredUserBodyFields,
    validateSameEmailDoesntExist,
    createUser,
  );

router
  .route(`/users/:userId`)
  .all(validateUserExists, onlySameUserOrAdminCanDoThisAction)
  .get(getUserById)
  .delete(removeUser)
  .put([
    body('email').isEmail(),
    body('password')
      .isLength({ min: 5 })
      .withMessage('Must include password (5+ characters)'),
    body('firstName').isString(),
    body('lastName').isString(),
    body('permissionLevel').isInt(),
    verifyBodyFieldsErrors,
    validateSameEmailBelongToSameUser,
    onlySameUserOrAdminCanDoThisAction,
    minimumPermissionLevelRequired(PermissionLevel.PaidPermission),
    put,
  ])
  .patch([
    body('email').isEmail().optional(),
    body('password')
      .isLength({ min: 5 })
      .withMessage('Password must be 5+ characters')
      .optional(),
    body('firstName').isString().optional(),
    body('lastName').isString().optional(),
    body('permissionLevel').isInt().optional(),
    verifyBodyFieldsErrors,
    validatePatchEmail,
    onlySameUserOrAdminCanDoThisAction,
    minimumPermissionLevelRequired(PermissionLevel.PaidPermission),
    patch,
  ]);

export default router;
