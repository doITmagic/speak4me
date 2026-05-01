#!/bin/bash
# =============================================================================
# CONVERSIE F5-TTS Romanian → ONNX (rulat O SINGURĂ DATĂ)
# După conversie, Python NU mai este necesar. Aplicația rulează nativ Go.
# =============================================================================
set -e

PROJECT_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
ONNX_OUTPUT="$PROJECT_ROOT/models/f5-tts-romanian"
CHECKPOINT="/tmp/f5-ro-model"
EXPORT_TOOL="/tmp/f5-export-tool"
VOCOS="/tmp/vocos-mel-24khz"
VENV="/tmp/f5conv/venv"

# System Python (nu conda!)
SYS_PYTHON=/usr/bin/python3.12

echo "============================================="
echo "  F5-TTS Romanian → ONNX (conversie unică)"
echo "============================================="
echo ""

# 1. Creem venv cu system Python
echo "📦 [1/5] Mediu Python izolat (system, nu conda)..."
if [ ! -d "$VENV" ]; then
    $SYS_PYTHON -m venv "$VENV"
fi
source "$VENV/bin/activate"
pip install --upgrade pip -q

# 2. Dependențe
echo "📦 [2/5] Dependențe de export..."
pip install f5-tts onnx onnxruntime huggingface_hub pydub pypinyin jieba omegaconf vocos -q 2>&1 | tail -1

# 3. Checkpoint românesc
echo "📦 [3/5] Checkpoint cdorob/f5-tts-romanian..."
if [ ! -f "$CHECKPOINT/model_last.pt" ]; then
    echo "   Descărcăm (5.4 GB)..."
    python3 -c "
from huggingface_hub import snapshot_download
snapshot_download(repo_id='cdorob/f5-tts-romanian', local_dir='$CHECKPOINT')
"
fi
echo "   ✅ $(ls -lh $CHECKPOINT/model_last.pt | awk '{print $5, $NF}')"

# 4. Vocos vocoder
echo "📦 [4/5] Vocos vocoder..."
if [ ! -d "$VOCOS" ]; then
    python3 -c "
from huggingface_hub import snapshot_download
snapshot_download(repo_id='charactr/vocos-mel-24khz', local_dir='$VOCOS')
"
fi
echo "   ✅ Vocos prezent"

# 5. Tool de export DakeQQ
echo "📦 [5/5] DakeQQ F5-TTS-ONNX export tool..."
if [ ! -d "$EXPORT_TOOL" ]; then
    git clone --depth 1 https://github.com/DakeQQ/F5-TTS-ONNX.git "$EXPORT_TOOL"
fi

# Export
echo ""
echo "🔧 EXPORTĂM MODELUL ÎN ONNX..."
echo "   Asta durează 5-15 minute (model mare)."
echo ""
mkdir -p "$ONNX_OUTPUT"

cd "$EXPORT_TOOL/Export_ONNX/F5_TTS"

python3 Export_F5.py \
    --f5safetensor_path "$CHECKPOINT/model_last.pt" \
    --vocab_path "$CHECKPOINT/vocab.txt" \
    --vocosmodel_dir "$VOCOS" \
    --preprocessmodel_path "$ONNX_OUTPUT/F5_Preprocess.onnx" \
    --transformermodel_path "$ONNX_OUTPUT/F5_Transformer.onnx" \
    --decodermodel_path "$ONNX_OUTPUT/F5_Decode.onnx" \
    --testlang en

echo ""
echo "============================================="
echo "✅ EXPORT COMPLET!"
echo ""
echo "Fișiere ONNX generate:"
ls -lh "$ONNX_OUTPUT/"*.onnx
echo ""
echo "Acum aplicația Go poate rula 100% nativ."
echo "Python nu mai este necesar."
echo "============================================="

deactivate
