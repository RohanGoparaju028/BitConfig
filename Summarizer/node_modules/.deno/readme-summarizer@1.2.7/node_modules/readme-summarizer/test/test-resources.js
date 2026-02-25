const readmeUrl = 'https://raw.githubusercontent.com/bennymeg/IsraelPostalServiceAPI/master/README.md';
const readmeUrlTest ='https://raw.githubusercontent.com/bennymeg/ReadmeSummarizer/master/README.md';

const expectedShortDescription = "An API for Israel postal service - query shipment price easily.";
const expectedShortTestDescription = "Summaries markdown readme files into few sentences.";

const expectedLongDescription = "An API for Israel postal service - query shipment price easily.\n" +
    "Supports both Node and browser application bases on browserify / webpack (including Angular).";
const expectedLongTestDescription = "Summaries markdown readme files into few sentences. " +
    "Get short description for any repository easily.";

const readme = `# Israel Postal Service API
An API for Israel postal service - query shipment price easily.
Supports both Node and browser application bases on browserify / webpack (including Angular).

[![licence](https://img.shields.io/github/license/mashape/apistatus.svg)](https://github.com/bennymeg/IsraelPostalServiceAPI/blob/master/LICENSE)
[![npm version](https://img.shields.io/npm/v/israel-postal-service-api.svg)](https://www.npmjs.com/package/israel-postal-service-api)
<!-- [![github version](https://img.shields.io/github/package-json/v/badges/shields.svg)](https://github.com/bennymeg/IsraelPostalServiceAPI) -->
<!-- ![GitHub repository size in bytes](https://img.shields.io/github/languages/code-size/badges/shields.svg) -->


## Installation
install...`;

const readmeTest = `# Readme Summarizer
Summaries markdown readme files into few sentences. Get short description for any repository easily.

[![licence](https://img.shields.io/github/license/mashape/apistatus.svg)](https://github.com/bennymeg/ReadmeSummarizer/blob/master/LICENSE)
[![npm version](https://img.shields.io/npm/v/readme-summarizer.svg)](https://www.npmjs.com/package/readme-summarizer)


## Installation
install...`;

module.exports.expectedShortDescription = expectedShortDescription;
module.exports.expectedShortTestDescription = expectedShortTestDescription;
module.exports.expectedLongDescription = expectedLongDescription;
module.exports.expectedLongTestDescription = expectedLongTestDescription;
module.exports.readme = readme;
module.exports.readmeTest = readmeTest;
module.exports.readmeUrl = readmeUrl;
module.exports.readmeUrlTest = readmeUrlTest;