# Readme Summarizer
Summaries markdown readme files into few sentences. Get short description for any repository easily.

[![licence](https://img.shields.io/github/license/bennymeg/ReadmeSummarizer.svg)](https://github.com/bennymeg/ReadmeSummarizer/blob/master/LICENSE)
[![npm version](https://img.shields.io/npm/v/readme-summarizer.svg)](https://www.npmjs.com/package/readme-summarizer)
[![Dependencies status](https://david-dm.org/bennymeg/ReadmeSummarizer/status.svg)](https://david-dm.org/bennymeg/ReadmeSummarizer)


## Installation
```bash
npm install --save readme-summarizer
```
## Usage

### Import library:
```javascript
import { ReadmeSummarizer } from 'readme-summarizer';
```

### Get readme summary from URL:
```javascript
let readmeUrl = "https://raw.githubusercontent.com/bennymeg/ReadmeSummarizer/master/README.md";
let shortSummary = true;

// get readme summary asynchronously from URL
ReadmeSummarizer.fromUrl(readmeUrl, shortSummary).then((response) => {
    console.log(response);
    // prints: "Summaries markdown readme files into few sentences."
}).catch((error) => {
    console.error('Error:', error);
});
```

### Get readme summary from text:

```javascript
let readmeAsText = "#Your markdown readme text string...";
let shortSummary = false;

// get readme summary from text string
ReadmeSummarizer.fromText(readmeUrl, shortSummary);
```
## Supported Environments ##
- 💻 Browser (including browserify / webpack based environments [such as Angular])
- 🖥  Node.js

## Documentation ##  
- 👨🏼‍💻 [API](https://github.com/bennymeg/ReadmeSummarizer/blob/master/docs/API.md),  
- 👩🏼‍🏫 [Examples](https://github.com/bennymeg/ReadmeSummarizer/blob/master/docs/examples),  
<!-- - 📜 [Change log](https://github.com/bennymeg/ReadmeSummarizer/blob/master/docs/CHANGELOG.md),   -->
- 🖋  [Licence](https://github.com/bennymeg/ReadmeSummarizer/blob/master/LICENSE)

## Support ##
If you're having any problem, please [raise an issue](https://github.com/bennymeg/ReadmeSummarizer/issues/new) on GitHub and we'll be happy to help.


## Contribute ##
- 👾 [Issue Tracker](https://github.com/bennymeg/ReadmeSummarizer/issues),
- 📦 [Source Code](https://github.com/bennymeg/ReadmeSummarizer/)

Any help generifying this repository will be very much appreciated. 
Before submitting a pull request, please make sure that you include tests, and that [jshint](http://jshint.com) runs without any warnings: [Download VSCode extension](https://marketplace.visualstudio.com/items?itemName=dbaeumer.jshint).

## Test ## 
Run the test suite by executing:

```sh
$ npm test
```

___

**Author:** Benny Megidish.
