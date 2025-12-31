#!/usr/bin/env python3
"""
Extract Welsh vocabulary from PDF wordlists and create a TSV file.
"""
import re
import subprocess
import sys
from pathlib import Path

def extract_text_from_pdf(pdf_path):
    """Extract text from PDF using pdftotext."""
    try:
        result = subprocess.run(
            ['pdftotext', '-layout', pdf_path, '-'],
            capture_output=True,
            text=True,
            check=True
        )
        return result.stdout
    except subprocess.CalledProcessError as e:
        print(f"Error extracting {pdf_path}: {e}", file=sys.stderr)
        return ""
    except FileNotFoundError:
        print("pdftotext not found. Please install poppler-utils.", file=sys.stderr)
        sys.exit(1)

def extract_vocab_pairs(text):
    """Extract Welsh-English word pairs from text."""
    pairs = []

    # Look for lines with Welsh word followed by English translation
    # Pattern matches: word(s) -------- translation
    # Also handles patterns like: word(suffix) -------- translation(s)

    lines = text.split('\n')
    in_geirfa = False

    for line in lines:
        # Check if we're in a geirfa section
        if 'Geirfa' in line or 'geirfa' in line.lower():
            in_geirfa = True
            continue

        # Skip section headers and category labels
        if any(x in line for x in ['enwau benywaidd', 'enwau gwrywaidd', 'berfau',
                                     'ansoddeiriau', 'arall', 'feminine nouns',
                                     'masculine nouns', 'verbs', 'adjectives', 'other',
                                     'Cymraeg', 'Welsh', 'Nod:', 'Uned']):
            continue

        # Stop at next section
        if in_geirfa and line.strip() and any(x in line for x in ['Sgwrs', 'Ymarfer', 'Deialog']):
            in_geirfa = False
            continue

        if in_geirfa:
            # Match patterns like: word -------- translation
            # or: word(suffix) -------- translation
            match = re.search(r'^([a-zâêîôûŵŷäëïöüẅÿáéíóúẃýàèìòùẁỳ\'\?\!]+(?:\([^)]+\))?)\s+[-—]+\s+(.+)$',
                            line.strip(), re.IGNORECASE)
            if match:
                welsh = match.group(1).strip()
                english = match.group(2).strip()
                # Clean up the English translation
                english = re.sub(r'\s+[-—]+.*$', '', english)  # Remove trailing dashes
                if welsh and english:
                    pairs.append((welsh, english))

    return pairs

def main():
    """Main function to process all PDFs and create TSV."""
    research_dir = Path('/Users/str20tbl/code/dysgair/dissertation/research')
    output_file = Path('/Users/str20tbl/code/dysgair/dissertation/research/vocab_wordlist.tsv')

    # Find all wordlist PDFs
    pdf_files = sorted(research_dir.glob('wordlist copy *.pdf'),
                      key=lambda x: int(re.search(r'(\d+)', x.stem).group(1)))

    print(f"Found {len(pdf_files)} PDF files")

    all_pairs = []
    seen = set()  # To avoid duplicates

    for pdf_file in pdf_files:
        print(f"Processing {pdf_file.name}...")
        text = extract_text_from_pdf(str(pdf_file))
        pairs = extract_vocab_pairs(text)

        # Add unique pairs
        for welsh, english in pairs:
            key = (welsh.lower(), english.lower())
            if key not in seen:
                seen.add(key)
                all_pairs.append((welsh, english))

        print(f"  Found {len(pairs)} vocabulary entries")

    # Write to TSV
    print(f"\nWriting {len(all_pairs)} unique vocabulary entries to {output_file}")
    with open(output_file, 'w', encoding='utf-8') as f:
        f.write("Welsh\tEnglish\n")
        for welsh, english in all_pairs:
            f.write(f"{welsh}\t{english}\n")

    print("Done!")

if __name__ == '__main__':
    main()
