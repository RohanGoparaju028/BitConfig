const dynamicFetch = require('./dynamic/node-fetch').fetch;

/**
 * Readme Summarizer
 * @author Benny Megidish
 */
class ReadmeSummarizer {
    constructor() {}
    
    /**
     * Summarize markdown readme file from a given URL
     * @param {string} url readme file URL.
     * @param {boolean} short get short summary (default: true).
     * @returns {Promise<string>} a promise to receive the readme summary.
     */
    static fromUrl(url, short=true) {
        return fetchReadmeFile(url)
            .then(response => response.text())
            .then(response => {
                return this.fromMarkdownText(response, short);
            }).catch(error => console.log(error));
    }

    /**
     * Summarize markdown readme file from a given text string
     * @param {string} text readme file as string.
     * @param {boolean} short get short summary (default: true).
     * @returns {Promise<string>} a promise to receive the readme summary.
     */
    static fromMarkdownText(text, short=true) {
        // convert titles to sections
        let cleanReadme = sectionizeMarkdown(String(text));

        // add section to the end of the file
        cleanReadme = cleanReadme.concat(`\n${SECTION}`);

        // trim all text between the two first headlines
        cleanReadme = cleanReadme.replace(new RegExp(`^${SECTION}(.+?)${SECTION}.*$`, 'gsm'), '$1').trim();

        // remove badges and comments
        cleanReadme = removeHtml(removeComments(removeBadges(cleanReadme)));

        // remove markdown from the description
        let longDescription = removeTables(cleanReadme);

        // get the first descriptive line
        let shortDescription = short ? getShortDescription(longDescription) : undefined;

        return removeEmptyLines(removeMarkdown(short ? shortDescription : longDescription));
    }
}

const SECTION = '#SECTION';

function getShortDescription(longDescription) {
    //let shortDescriptionMatch = longDescription.match(new RegExp('^(.+)(\n|\.\s+).*$', 'm'));
    //let shortDescription = shortDescriptionMatch ? shortDescriptionMatch[1] : longDescription;
    let shortDescriptionMatch = longDescription.replace(/(\n+|(\.)\s+)/g, '$2. ').split(/\.\s/);

    // find the first descriptive line
    let index = shortDescriptionMatch.findIndex((line) => {
        return !line.match(/^\s*\[!\[(.*?)\]\(.*?\)\]\(.*?\)\s*$|^\s*!?\[(.*?)\]\(.*?\)\s*$/gm);
    });

    let shortDescription = shortDescriptionMatch[index > 0 ? index : 0];

    return shortDescription;
}

function fetchReadmeFile(url) {
    return dynamicFetch(url); 
}

function removeBadges(text) {
    // remove all badges which ar not part of a sentence
    return text.replace(/^\s*\[!.*\]\(.*\)|^\s*!\[(.*?)\]\(.*?\)\s*$/gm, ''); 
}

function removeComments(text) {
    // remove comments from the text
    return text.replace(/<!--.*-->/gm, ''); 
}

function removeHtml(text) {
    // remove html markdown from the text
    return text.replace(/<[^!][^>]*>/gm, ''); 
}

function removeEmptyLines(text) {
    // remove all the empty lines from the text
    return text.replace(/^\s*[\r\n]/gm, '').trim('\n').trim('\r').trim(); 
}

function removeTables(text) {
    // strip the text from markdown tables
    return text.replace(/^\|?.+\|.+\|.+\|?.*$/gm, ''); 
}

function removeMarkdown(text) {
    // strip the text from markdown symbols, convert link and inline badges to text
    return text.replace(/(?:__|[*#`])|\[!\[(.*?)\]\(.*?\)\]\(.*?\)|!?\[(.*?)\]\(.*?\)/gm, '$1$2'); 
}

function sectionizeMarkdown(text) {
    // convert all the titles and subtitles to known section title

    // convert --- or === titles to regular # titles
    text = text.replace(/^([A-Za-z\s]+)(?:-{3,}|={3,})$/gm, '# $1 \n');

    // handle # titles
    return text.replace(/^#+.*$/gm, SECTION); 
}

module.exports.ReadmeSummarizer = ReadmeSummarizer;