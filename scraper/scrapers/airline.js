import { Chalk } from 'chalk';

// Environment variables
let { IS_PROVISIONER } = process.env;

let colorLevel = 1;

if (IS_PROVISIONER) {
    colorLevel = 0;
}
const customChalk = new Chalk({ level: colorLevel });

