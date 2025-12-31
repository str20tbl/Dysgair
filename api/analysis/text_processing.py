"""
Welsh Text Processing Module
Handles text normalization, tokenization, and character classification for Welsh language
"""

import unicodedata
from typing import List, Tuple


# Welsh character classification constants
WELSH_DIGRAPHS = ['ll', 'ch', 'dd', 'ff', 'ng', 'rh', 'ph', 'th']
WELSH_VOWELS = set('aeiouyẃâêîôûŵŷàèìòùẁỳäëïöüẅÿ')  # includes all diacritics
WELSH_CONSONANTS = set('bcdfghjklmnpqrstvxz')


class WelshTextProcessor:
    """
    Handles text processing operations specific to Welsh language analysis.

    Provides normalization, tokenization, and character classification
    tailored for Welsh orthography and ASR evaluation.
    """

    @staticmethod
    def normalize_text(text: str) -> str:
        """
        Normalize text for consistent metric calculations by:
          - Converting to lowercase
          - Replacing hyphens with spaces (e.g., "well-known" → "well known")
          - Removing all Unicode punctuation and symbols (e.g., .,!?;:"' … — – '' "")
          - Normalizing whitespace (trim and collapse multiple spaces)

        This ensures fair comparison between ASR models, particularly when one model
        (Whisper) produces punctuation while another (Wav2Vec2) does not.

        Args:
            text: Text to normalize

        Returns:
            Normalized text string
        """
        # Convert to lowercase
        text = text.lower()

        # Replace hyphens with spaces to split hyphenated words
        text = text.replace("-", " ")

        # Remove all Unicode punctuation and symbols
        text = ''.join(
            char for char in text
            if not (unicodedata.category(char).startswith('P') or
                   unicodedata.category(char).startswith('S'))
        )

        # Normalize whitespace: trim and collapse multiple spaces
        text = ' '.join(text.split())

        return text

    @staticmethod
    def apply_lenient_normalization(target: str, transcription: str) -> Tuple[str, str]:
        """
        Apply lenient space-insensitive matching between target and transcription.

        This function checks if the target text (with all spaces removed) exactly matches
        the transcription (with all spaces removed). For single-word targets, it also
        matches if the transcription's first word matches the target (ignoring over-transcription).

        This approach gives transcriptions benefit when spacing differences or ASR
        over-transcription are the only errors, while still catching actual character-level mistakes.

        Args:
            target: Normalized target text (should already be normalized)
            transcription: Normalized transcription text (should already be normalized)

        Returns:
            Tuple of (target, transcription) where both are set to target if lenient match found,
            otherwise returns original values

        Examples:
            - Target "hello world" matches transcription "hel lo wor ld" → ("hello world", "hello world")
            - Target "rhyw" matches transcription "rhyw reswm wrth..." → ("rhyw", "rhyw") [first word match]
            - Target "rhyw" does NOT match "rhywbeth" → ("rhyw", "rhywbeth") [different words]
            - Target "test" doesn't match "demo" → ("test", "demo")
        """
        # Remove all spaces from both strings for comparison
        target_no_spaces = target.replace(' ', '')
        transcription_no_spaces = transcription.replace(' ', '')

        # Check for EXACT MATCH after removing spaces (not substring!)
        if target_no_spaces and target_no_spaces == transcription_no_spaces:
            # Exact match found! Return target for both to get perfect metrics
            return target, target

        # For single-word targets, check if transcription's first word matches
        # This handles ASR over-transcription (e.g., "rhyw" → "rhyw reswm wrth...")
        if target and ' ' not in target.strip():
            # Target is a single word
            transcription_words = transcription.split()
            if transcription_words and transcription_words[0] == target:
                # First word of transcription matches target exactly
                return target, target

        # No match - return original values
        return target, transcription

    @staticmethod
    def tokenize_welsh(text: str) -> List[str]:
        """
        Tokenize Welsh text into letters and digraphs.

        Welsh digraphs (ll, ch, dd, ff, ng, rh, ph, th) are treated as single units.

        Args:
            text: Welsh text to tokenize (should be normalized first)

        Returns:
            List of tokens (individual letters or digraph markers like ⟨ll⟩)
        """
        # Normalize the text first
        text = text.lower().strip()

        # Replace digraphs with unique markers
        for digraph in WELSH_DIGRAPHS:
            text = text.replace(digraph, f'⟨{digraph}⟩')

        # Split into individual characters/tokens
        tokens = []
        i = 0
        while i < len(text):
            # Check if this is a digraph marker
            if i < len(text) - 1 and text[i] == '⟨':
                # Find the closing ⟩
                end = text.find('⟩', i)
                if end != -1:
                    tokens.append(text[i:end+1])
                    i = end + 1
                    continue

            # Regular character
            if text[i] not in ' \t\n':  # Skip whitespace
                tokens.append(text[i])
            i += 1

        return tokens

    @staticmethod
    def classify_token(token: str) -> str:
        """
        Classify a token as vowel, consonant, digraph, or other.

        Args:
            token: Single character or digraph marker (e.g., 'a', 'b', '⟨ll⟩')

        Returns:
            Category: "vowel", "consonant", "digraph", or "other"
        """
        # Check if it's a digraph marker
        if token.startswith('⟨') and token.endswith('⟩'):
            return "digraph"

        # Single character classification
        if len(token) == 1:
            if token in WELSH_VOWELS:
                return "vowel"
            elif token in WELSH_CONSONANTS:
                return "consonant"

        return "other"
