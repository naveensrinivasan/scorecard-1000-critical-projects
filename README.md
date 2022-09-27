# scorecard-1000-critical-projects

The goal of this is to demonstrate how to use the [scorecard API](https://github.com/ossf/scorecard) to analyze 
the top 1000 critical projects on GitHub. The list of projects is taken from the [Criticality Score](https://github.com/ossf/criticality_score) project.

## How to run
`go run main.go`

## Data
- [results folder](./results)
- [criticality_score.csv](all.csv)

## Why?

The criticality score is a great way to identify the most important projects on GitHub. 
But the consumers of those projects aren't aware of the security and quality of the projects they are using.
This provides an easy way to see the security and quality of the top 1000 projects on GitHub.


## What's next?

- [ ] Run this periodically to track the security and quality of the top 1000 projects on GitHub.
- [ ] Publish the results that are accessible to the public as a REST endpoint (GCP buckets are the easiest way to do this.) 
- [ ] Create a dashboard to visualize the results. - Need help with this.
- [ ] Extend it to use also the harvard census dataset. 