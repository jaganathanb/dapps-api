import swaggerAutogen from 'swagger-autogen';
import swaggerDef from './swagger.def';

const outputFile = './swagger-output.json';
const endpointsFiles = ['./src/users/users.routes.ts'];

swaggerAutogen()(outputFile, endpointsFiles, swaggerDef);
