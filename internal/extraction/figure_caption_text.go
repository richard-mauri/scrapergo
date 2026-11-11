package extraction

import (
    "clean-article-reader/internal/article"
    "clean-article-reader/internal/sourcehtml"
)

// removeFlattenedFigureCaptions reconciles Readability's text representation
// against the figure semantics in its own selected HTML. RenderText may emit
// figcaption text as independent paragraphs; the image block added later by
// PositionArticleFigures already owns that same caption.
//
// Only source-backed captions belonging to figures with actual images qualify.
// Any matching standalone editorial prose in the selected HTML protects the
// text from removal. This deliberately does not deduplicate general prose.
func removeFlattenedFigureCaptions(blocks []article.Block, selectedHTML string) []article.Block {
    if len(blocks) == 0 || selectedHTML == "" {
        return blocks
    }
    root := sourcehtml.Parse(selectedHTML)
    captionText := map[string]bool{}
    for _, fig := range sourcehtml.Find(root, func(n *sourcehtml.Node) bool { return n.Tag == "figure" }) {
        if len(sourcehtml.Find(fig, func(n *sourcehtml.Node) bool { return n.Tag == "img" })) == 0 {
            continue
        }
        for _, cap := range sourcehtml.Find(fig, func(n *sourcehtml.Node) bool { return n.Tag == "figcaption" }) {
            if t := sourcehtml.Normalize(sourcehtml.Text(cap)); t != "" {
                captionText[t] = true
            }
            // Figure captions commonly contain a descriptive span and a
            // separate credit span. RenderText can emit both independently
            // and also emit their joined parent text.
            for _, child := range cap.Kids {
                if child.Tag == "#text" {
                    continue
                }
                if t := sourcehtml.Normalize(sourcehtml.Text(child)); t != "" {
                    captionText[t] = true
                }
            }
        }
    }
    if len(captionText) == 0 {
        return blocks
    }
    // A separate article paragraph with identical words is NOT a redundant
    // caption. Preserve those words even when the figure uses them too.
    for _, n := range sourcehtml.Find(root, func(n *sourcehtml.Node) bool {
        return n.Tag == "p" || n.Tag == "blockquote" || n.Tag == "li" || n.Tag == "h2" || n.Tag == "h3" || n.Tag == "h4"
    }) {
        insideCaption := false
        for parent := n.Parent; parent != nil; parent = parent.Parent {
            if parent.Tag == "figcaption" {
                insideCaption = true
                break
            }
        }
        if !insideCaption {
            delete(captionText, sourcehtml.Normalize(sourcehtml.Text(n)))
        }
    }
    kept := make([]article.Block, 0, len(blocks))
    for _, block := range blocks {
        if block.Kind == "p" && captionText[sourcehtml.Normalize(block.Text)] {
            continue
        }
        kept = append(kept, block)
    }
    return kept
}
