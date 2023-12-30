import path from 'path';

import dotenv from 'dotenv';
import Joi from 'joi';

type ENVVARS = {
  NODE_ENV: string;
  PORT: number;
  JWT_SECRET: string;
  JWT_ACCESS_EXPIRATION_MINUTES: number;
  JWT_REFRESH_EXPIRATION_DAYS: number;
  JWT_RESET_PASSWORD_EXPIRATION_MINUTES: number;
  JWT_VERIFY_EMAIL_EXPIRATION_MINUTES: number;
  SMTP_HOST: string;
  SMTP_PORT: number;
  SMTP_USERNAME: string;
  SMTP_PASSWORD: string;
  EMAIL_FROM: string;
};

dotenv.config({ path: path.join(__dirname, '../../.env') });

const envVarsSchema = Joi.object<ENVVARS>()
  .keys({
    NODE_ENV: Joi.string()
      .valid('production', 'development', 'test')
      .required(),
    PORT: Joi.number().default(7770),
    JWT_SECRET: Joi.string().required().description('JWT secret key'),
    JWT_ACCESS_EXPIRATION_MINUTES: Joi.number()
      .default(30)
      .description('minutes after which access tokens expire'),
    JWT_REFRESH_EXPIRATION_DAYS: Joi.number()
      .default(30)
      .description('days after which refresh tokens expire'),
    JWT_RESET_PASSWORD_EXPIRATION_MINUTES: Joi.number()
      .default(10)
      .description('minutes after which reset password token expires'),
    JWT_VERIFY_EMAIL_EXPIRATION_MINUTES: Joi.number()
      .default(10)
      .description('minutes after which verify email token expires'),
    SMTP_HOST: Joi.string().description('server that will send the emails'),
    SMTP_PORT: Joi.number().description('port to connect to the email server'),
    SMTP_USERNAME: Joi.string().description('username for email server'),
    SMTP_PASSWORD: Joi.string().description('password for email server'),
    EMAIL_FROM: Joi.string().description(
      'the from field in the emails sent by the app',
    ),
  })
  .unknown();

const result = envVarsSchema
  .prefs({ errors: { label: 'key' } })
  .validate(process.env);

if (result.error) {
  throw new Error(`Config validation error: ${result.error.message}`);
}

const envVars = result.value;

const defaultConfig = {
  env: envVars.NODE_ENV,
  port: envVars.PORT,
  jwt: {
    secret: envVars.JWT_SECRET,
    accessExpirationMinutes: envVars.JWT_ACCESS_EXPIRATION_MINUTES,
    refreshExpirationDays: envVars.JWT_REFRESH_EXPIRATION_DAYS,
    resetPasswordExpirationMinutes:
      envVars.JWT_RESET_PASSWORD_EXPIRATION_MINUTES,
    verifyEmailExpirationMinutes: envVars.JWT_VERIFY_EMAIL_EXPIRATION_MINUTES,
  },
  email: {
    smtp: {
      host: envVars.SMTP_HOST,
      port: envVars.SMTP_PORT,
      auth: {
        user: envVars.SMTP_USERNAME,
        pass: envVars.SMTP_PASSWORD,
      },
    },
    from: envVars.EMAIL_FROM,
  },
};

export { defaultConfig };
