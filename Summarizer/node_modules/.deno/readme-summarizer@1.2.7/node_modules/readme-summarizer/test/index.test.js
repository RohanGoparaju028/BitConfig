const ReadmeSummarizer = require('../index').ReadmeSummarizer;
const testResources = require('./test-resources');
const assert = require('chai').assert;

describe('Readme Summarizer', () => {
    let expectedShortDescription = testResources.expectedShortDescription;
    let expectedLongDescription = testResources.expectedLongDescription;

    describe('From text', () => {
        let readme = testResources.readme;

        it('should summarize readme to a long description from text', () => {
            let description = ReadmeSummarizer.fromMarkdownText(readme, false);
            assert.equal(description, expectedLongDescription);
        });

        it('should summarize readme to a short description from text', () => {
            let description = ReadmeSummarizer.fromMarkdownText(readme, true);
            assert.equal(description, expectedShortDescription);
        });
    });

    describe('From URL', () => {
        let url = testResources.readmeUrl;

        it('should summarize readme to a long description from url', async () => {
            const result = await ReadmeSummarizer.fromUrl(url, false);
            assert.equal(result, expectedLongDescription);
        });

        it('should summarize readme to a short description from url', async () => {
            const result = await ReadmeSummarizer.fromUrl(url, true);
            assert.equal(result, expectedShortDescription);
        });
    });
});